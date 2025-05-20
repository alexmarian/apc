package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type ExpenseItem struct {
	ID             int64     `json:"id,omitempty"`
	Amount         float64   `json:"amount,omitempty"`
	Description    string    `json:"description,omitempty"`
	Destination    string    `json:"destination,omitempty"`
	Date           time.Time `json:"date,omitempty"`
	Month          int64     `json:"month,omitempty"`
	Year           int64     `json:"year,omitempty"`
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

func HandleCreateExpense(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var expense struct {
			Amount      float64   `json:"amount"`
			Description string    `json:"description"`
			Destination string    `json:"destination"`
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
		if expense.Amount <= 0 {
			RespondWithError(rw, http.StatusBadRequest, "Amount must be positive")
			return
		}

		if expense.Description == "" {
			RespondWithError(rw, http.StatusBadRequest, "Description is required")
			return
		}

		// Verify category and account belong to the association
		category, err := cfg.Db.GetCategory(req.Context(), expense.CategoryID)
		if err != nil || category.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusBadRequest, "Invalid category")
			return
		}

		account, err := cfg.Db.GetAccount(req.Context(), expense.AccountID)
		if err != nil || account.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusBadRequest, "Invalid account")
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

		// Return created expense
		RespondWithJSON(rw, http.StatusCreated, ExpenseItem{
			ID:          newExpense.ID,
			Amount:      newExpense.Amount,
			Description: newExpense.Description,
			Destination: newExpense.Destination,
			Date:        newExpense.Date,
			Month:       newExpense.Month,
			Year:        newExpense.Year,
			CategoryID:  newExpense.CategoryID,
			AccountID:   newExpense.AccountID,
			CreatedAt:   newExpense.CreatedAt.Time,
			UpdatedAt:   newExpense.CreatedAt.Time,
		})
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
			Date:        existingExpense.Date,
			Month:       existingExpense.Month,
			Year:        existingExpense.Year,
			CategoryID:  existingExpense.CategoryID,
			AccountID:   existingExpense.AccountID,
		}

		// Apply updates
		if updateReq.Amount != nil {
			if *updateReq.Amount <= 0 {
				RespondWithError(rw, http.StatusBadRequest, "Amount must be positive")
				return
			}
			updateParams.Amount = *updateReq.Amount
		}

		if updateReq.Description != nil {
			if *updateReq.Description == "" {
				RespondWithError(rw, http.StatusBadRequest, "Description cannot be empty")
				return
			}
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

		if updateReq.CategoryID != nil {
			// Verify category belongs to association
			category, err := cfg.Db.GetCategory(req.Context(), *updateReq.CategoryID)
			if err != nil || category.AssociationID != int64(associationId) {
				RespondWithError(rw, http.StatusBadRequest, "Invalid category")
				return
			}
			updateParams.CategoryID = *updateReq.CategoryID
		}

		if updateReq.AccountID != nil {
			// Verify account belongs to association
			account, err := cfg.Db.GetAccount(req.Context(), *updateReq.AccountID)
			if err != nil || account.AssociationID != int64(associationId) {
				RespondWithError(rw, http.StatusBadRequest, "Invalid account")
				return
			}
			updateParams.AccountID = *updateReq.AccountID
		}

		// Update expense
		updatedExpense, err := cfg.Db.UpdateExpense(req.Context(), updateParams)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating expense", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update expense")
			return
		}

		// Return updated expense
		RespondWithJSON(rw, http.StatusOK, ExpenseItem{
			ID:          updatedExpense.ID,
			Amount:      updatedExpense.Amount,
			Description: updatedExpense.Description,
			Destination: updatedExpense.Destination,
			Date:        updatedExpense.Date,
			Month:       updatedExpense.Month,
			Year:        updatedExpense.Year,
			CategoryID:  updatedExpense.CategoryID,
			AccountID:   updatedExpense.AccountID,
			CreatedAt:   updatedExpense.CreatedAt.Time,
			UpdatedAt:   updatedExpense.CreatedAt.Time,
		})
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

		expensesItems := make([]ExpenseItem, len(expenses))

		for i, exp := range expenses {
			expensesItems[i] = ExpenseItem{
				ID:             exp.ID,
				Amount:         exp.Amount,
				Description:    exp.Description,
				Destination:    exp.Destination,
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
