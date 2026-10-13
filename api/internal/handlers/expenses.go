package handlers

import (
	"database/sql"
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

type ExpenseItem struct {
	ID             int64     `json:"id,omitempty"`
	Amount         float64   `json:"amount,omitempty"`
	Description    string    `json:"description,omitempty"`
	Destination    string    `json:"destination,omitempty"`
	Date           time.Time `json:"date,omitempty"`
	Month          int64     `json:"month,omitempty"`
	Year           int64     `json:"year,omitempty"`
	DocumentRef    string    `json:"document_ref,omitempty"`
	CategoryID     int64     `json:"category_id,omitempty"`
	CategoryType   string    `json:"category_type,omitempty"`
	CategoryFamily string    `json:"category_family,omitempty"`
	CategoryName   string    `json:"category_name,omitempty"`
	AccountID      int64     `json:"account_id,omitempty"`
	AccountNumber  string    `json:"account_number,omitempty"`
	AccountName    string    `json:"account_name,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

const ExpensesIdPathValue = "expenseId"

const (
	expenseDescriptionMax = 255
	expenseDestinationMax = 255
	expenseDocumentRefMax = 50
)

// validateExpenseText trims the free-text fields in place and checks their
// limits; returns a message for a 400, or "".
func validateExpenseText(description, destination, documentRef *string) string {
	for _, f := range []*string{description, destination, documentRef} {
		if f != nil {
			*f = strings.TrimSpace(*f)
		}
	}
	switch {
	case description != nil && *description == "":
		return "Description is required"
	case description != nil && utf8.RuneCountInString(*description) > expenseDescriptionMax:
		return "Description must be at most 255 characters"
	case destination != nil && utf8.RuneCountInString(*destination) > expenseDestinationMax:
		return "Destination must be at most 255 characters"
	case documentRef != nil && utf8.RuneCountInString(*documentRef) > expenseDocumentRefMax:
		return "Document reference must be at most 50 characters"
	}
	return ""
}

func validAmount(a float64) bool { return a > 0 && !math.IsInf(a, 0) && !math.IsNaN(a) }

// expenseCategory checks a category belongs to the association and, for a
// new assignment, is active. On failure it has already written the 400.
func expenseCategory(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, associationId, categoryId int64) bool {
	category, err := cfg.Db.GetCategory(req.Context(), categoryId)
	if err != nil || category.AssociationID != associationId {
		RespondWithError(rw, http.StatusBadRequest, "Invalid category")
		return false
	}
	if category.IsDeleted {
		RespondWithError(rw, http.StatusBadRequest, "The category is inactive")
		return false
	}
	return true
}

// expenseAccount: as expenseCategory, for accounts (disabled ones refused).
func expenseAccount(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, associationId, accountId int64) bool {
	account, err := cfg.Db.GetAccount(req.Context(), accountId)
	if err != nil || account.AssociationID != associationId {
		RespondWithError(rw, http.StatusBadRequest, "Invalid account")
		return false
	}
	if !account.IsActive {
		RespondWithError(rw, http.StatusBadRequest, "The account is disabled")
		return false
	}
	return true
}

// respondWithExpense answers with the expense joined with its category and
// account (the shape the list returns).
func respondWithExpense(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request, status int, associationId, expenseId int64) {
	e, err := cfg.Db.GetExpenseWithJoins(req.Context(), database.GetExpenseWithJoinsParams{ID: expenseId, AssociationID: associationId})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error fetching expense with joins", zap.Error(err))
		RespondWithError(rw, http.StatusInternalServerError, "Failed to fetch expense")
		return
	}
	RespondWithJSON(rw, status, ExpenseItem{
		ID: e.ID, Amount: e.Amount, Description: e.Description, Destination: e.Destination,
		DocumentRef: e.DocumentRef.String, Date: e.Date, Month: e.Month, Year: e.Year,
		CategoryID: e.CategoryID, CategoryType: e.CategoryType, CategoryFamily: e.CategoryFamily, CategoryName: e.CategoryName,
		AccountID: e.AccountID, AccountNumber: e.AccountNumber, AccountName: e.AccountName,
	})
}

func HandleCreateExpense(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var expense struct {
			Amount      float64   `json:"amount"`
			Description string    `json:"description"`
			Destination string    `json:"destination"`
			DocumentRef string    `json:"document_ref"`
			Date        time.Time `json:"date"`
			CategoryID  int64     `json:"category_id"`
			AccountID   int64     `json:"account_id"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&expense); err != nil {
			logging.Logger.Log(zap.WarnLevel, "Exception decoding expense", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate
		if !validAmount(expense.Amount) {
			RespondWithError(rw, http.StatusBadRequest, "Amount must be positive")
			return
		}
		if msg := validateExpenseText(&expense.Description, &expense.Destination, &expense.DocumentRef); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		if expense.Date.IsZero() {
			RespondWithError(rw, http.StatusBadRequest, "Date is required")
			return
		}

		// The category and account must belong to the association and be in use.
		if !expenseCategory(cfg, rw, req, int64(associationId), expense.CategoryID) ||
			!expenseAccount(cfg, rw, req, int64(associationId), expense.AccountID) {
			return
		}

		// Extract month and year from the date
		month := expense.Date.Month()
		year := expense.Date.Year()

		// Create expense
		newExpense, err := cfg.Db.CreateExpense(req.Context(), database.CreateExpenseParams{
			Amount:      expense.Amount,
			Description: expense.Description,
			Destination: expense.Destination,
			DocumentRef: sql.NullString{String: expense.DocumentRef, Valid: expense.DocumentRef != ""},
			Date:        expense.Date,
			Month:       int64(month),
			Year:        int64(year),
			CategoryID:  expense.CategoryID,
			AccountID:   expense.AccountID,
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating expense", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create expense")
			return
		}

		respondWithExpense(cfg, rw, req, http.StatusCreated, int64(associationId), newExpense.ID)
	}
}

// Update an expense
func HandleUpdateExpense(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		expenseId, _ := strconv.Atoi(req.PathValue(ExpensesIdPathValue))

		// Get existing expense to verify it exists and belongs to this association
		existingExpense, err := cfg.Db.GetExpenseWithAssociation(req.Context(), database.GetExpenseWithAssociationParams{
			ID:            int64(expenseId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Expense not found or doesn't belong to this association")
			} else {
				logging.Logger.Log(zap.WarnLevel, "Error retrieving expense", zap.Error(err))
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve expense")
			}
			return
		}

		// Parse update request
		type UpdateExpenseRequest struct {
			Amount      *float64   `json:"amount,omitempty"`
			Description *string    `json:"description,omitempty"`
			Destination *string    `json:"destination,omitempty"`
			DocumentRef *string    `json:"document_ref,omitempty"`
			Date        *time.Time `json:"date,omitempty"`
			CategoryID  *int64     `json:"category_id,omitempty"`
			AccountID   *int64     `json:"account_id,omitempty"`
		}

		var updateReq UpdateExpenseRequest
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&updateReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Prepare update params with existing values
		updateParams := database.UpdateExpenseParams{
			ID:          int64(expenseId),
			Amount:      existingExpense.Amount,
			Description: existingExpense.Description,
			Destination: existingExpense.Destination,
			DocumentRef: sql.NullString{String: existingExpense.DocumentRef.String, Valid: existingExpense.DocumentRef.Valid},
			Date:        existingExpense.Date,
			Month:       existingExpense.Month,
			Year:        existingExpense.Year,
			CategoryID:  existingExpense.CategoryID,
			AccountID:   existingExpense.AccountID,
		}

		// Apply updates
		if msg := validateExpenseText(updateReq.Description, updateReq.Destination, updateReq.DocumentRef); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		if updateReq.Amount != nil {
			if !validAmount(*updateReq.Amount) {
				RespondWithError(rw, http.StatusBadRequest, "Amount must be positive")
				return
			}
			updateParams.Amount = *updateReq.Amount
		}

		if updateReq.Description != nil {
			updateParams.Description = *updateReq.Description
		}

		if updateReq.Destination != nil {
			updateParams.Destination = *updateReq.Destination
		}

		if updateReq.Date != nil {
			updateParams.Date = *updateReq.Date
			updateParams.Month = int64(updateReq.Date.Month())
			updateParams.Year = int64(updateReq.Date.Year())
		}

		// A changed category/account must be in use; keeping the current one
		// is fine even if it has since been deactivated.
		if updateReq.CategoryID != nil && *updateReq.CategoryID != existingExpense.CategoryID {
			if !expenseCategory(cfg, rw, req, int64(associationId), *updateReq.CategoryID) {
				return
			}
			updateParams.CategoryID = *updateReq.CategoryID
		}

		if updateReq.AccountID != nil && *updateReq.AccountID != existingExpense.AccountID {
			if !expenseAccount(cfg, rw, req, int64(associationId), *updateReq.AccountID) {
				return
			}
			updateParams.AccountID = *updateReq.AccountID
		}
		if updateReq.DocumentRef != nil {
			updateParams.DocumentRef = sql.NullString{String: *updateReq.DocumentRef, Valid: *updateReq.DocumentRef != ""}
		} else {
			updateParams.DocumentRef = sql.NullString{String: existingExpense.DocumentRef.String, Valid: existingExpense.DocumentRef.Valid}
		}

		// Update expense
		if _, err := cfg.Db.UpdateExpense(req.Context(), updateParams); err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating expense", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update expense")
			return
		}

		// Same shape as the list and create: with category and account names.
		respondWithExpense(cfg, rw, req, http.StatusOK, int64(associationId), int64(expenseId))
	}
}

func HandleDeleteExpense(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		expenseId, _ := strconv.Atoi(req.PathValue(ExpensesIdPathValue))

		// Verify expense exists and belongs to association
		_, err := cfg.Db.GetExpenseWithAssociation(req.Context(), database.GetExpenseWithAssociationParams{
			ID:            int64(expenseId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Expense not found or doesn't belong to this association")
			} else {
				logging.Logger.Log(zap.WarnLevel, "Error retrieving expense", zap.Error(err))
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve expense")
			}
			return
		}

		// Delete expense
		err = cfg.Db.DeleteExpense(req.Context(), int64(expenseId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error deleting expense", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete expense")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Expense deleted successfully",
		})
	}
}
func HandleGetExpense(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		expenseId, _ := strconv.Atoi(req.PathValue(ExpensesIdPathValue))

		// Verify expense exists and belongs to association
		dbExpense, err := cfg.Db.GetExpenseWithAssociation(req.Context(), database.GetExpenseWithAssociationParams{
			ID:            int64(expenseId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Expense not found or doesn't belong to this association")
			} else {
				logging.Logger.Log(zap.WarnLevel, "Error retrieving expense", zap.Error(err))
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve expense")
			}
			return
		}

		RespondWithJSON(rw, http.StatusOK, ExpenseItem{
			ID:          dbExpense.ID,
			Amount:      dbExpense.Amount,
			Description: dbExpense.Description,
			Destination: dbExpense.Destination,
			DocumentRef: dbExpense.DocumentRef.String,
			Date:        dbExpense.Date,
			Month:       dbExpense.Month,
			Year:        dbExpense.Year,
			CategoryID:  dbExpense.CategoryID,
			AccountID:   dbExpense.AccountID,
			CreatedAt:   dbExpense.CreatedAt.Time,
			UpdatedAt:   dbExpense.CreatedAt.Time,
		})
	}
}
func HandleGetExpenseDestinations(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		destinations, err := cfg.Db.GetDistinctDestinations(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error fetching destinations", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to fetch destinations")
			return
		}

		if destinations == nil {
			destinations = []string{}
		}

		RespondWithJSON(rw, http.StatusOK, destinations)
	}
}

func HandleGetExpenses(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		startDate, endDate, err := GetRequestDateRange(req, &rw)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, err.Error())
			return
		}

		expenses, err := cfg.Db.GetExpensesByDateRange(req.Context(), database.GetExpensesByDateRangeParams{
			AssociationID: int64(associationId),
			Date:          startDate,
			Date_2:        endDate,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error listing expenses", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to list expenses")
			return
		}

		expensesItems := make([]ExpenseItem, len(expenses))

		for i, exp := range expenses {
			expensesItems[i] = ExpenseItem{
				ID:             exp.ID,
				Amount:         exp.Amount,
				Description:    exp.Description,
				Destination:    exp.Destination,
				DocumentRef:    exp.DocumentRef.String,
				Date:           exp.Date,
				CategoryID:     exp.CategoryID,
				CategoryType:   exp.CategoryType,
				CategoryFamily: exp.CategoryFamily,
				CategoryName:   exp.CategoryName,
				AccountID:      exp.AccountID,
				AccountNumber:  exp.AccountNumber,
				AccountName:    exp.AccountName,
			}
		}

		RespondWithJSON(rw, http.StatusOK, expensesItems)
	}
}
