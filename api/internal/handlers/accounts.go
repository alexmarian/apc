package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Account struct {
	ID            int64     `json:"id"`
	Number        string    `json:"number"`
	Destination   string    `json:"destination"`
	Description   string    `json:"description"`
	AssociationID int64     `json:"association_id"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

const AccountIdPathValue = "accountId"

func HandleGetAssociationAccounts(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		accounts, err := cfg.Db.GetAssociationAccounts(req.Context(), int64(associationId))
		if err != nil {
			var errorMsg = fmt.Sprintf("Error getting accounts: %s", err)
			log.Printf(errorMsg)
			RespondWithError(rw, http.StatusInternalServerError, errorMsg)
			return
		}

		// Convert database accounts to response format
		accountsResponse := make([]Account, len(accounts))
		for i, account := range accounts {
			accountsResponse[i] = Account{
				ID:            account.ID,
				Number:        account.Number,
				Destination:   account.Destination,
				Description:   account.Description,
				AssociationID: account.AssociationID,
				IsActive:      account.IsActive,
				CreatedAt:     account.CreatedAt.Time,
				UpdatedAt:     account.UpdatedAt.Time,
			}
		}

		RespondWithJSON(rw, http.StatusOK, accountsResponse)
	}
}

func HandleGetAccount(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		accountId, _ := strconv.Atoi(req.PathValue(AccountIdPathValue))

		// First get the account to verify it exists and belongs to the association
		account, err := cfg.Db.GetAccount(req.Context(), int64(accountId))
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Account not found")
			} else {
				log.Printf("Error getting account: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve account")
			}
			return
		}

		// Verify the account belongs to the association
		if account.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusForbidden, "Account does not belong to this association")
			return
		}

		// Convert to response format
		accountResponse := Account{
			ID:            account.ID,
			Number:        account.Number,
			Destination:   account.Destination,
			Description:   account.Description,
			AssociationID: account.AssociationID,
			IsActive:      account.IsActive,
			CreatedAt:     account.CreatedAt.Time,
			UpdatedAt:     account.UpdatedAt.Time,
		}

		RespondWithJSON(rw, http.StatusOK, accountResponse)
	}
}

func HandleCreateAccount(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		var accountRequest struct {
			Number      string `json:"number"`
			Destination string `json:"destination"`
			Description string `json:"description"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&accountRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate request
		if accountRequest.Number == "" {
			RespondWithError(rw, http.StatusBadRequest, "Account number is required")
			return
		}

		if accountRequest.Description == "" {
			RespondWithError(rw, http.StatusBadRequest, "Description is required")
			return
		}

		// Create account
		account, err := cfg.Db.CreateAccount(req.Context(), database.CreateAccountParams{
			Number:        accountRequest.Number,
			Destination:   accountRequest.Destination,
			Description:   accountRequest.Description,
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error creating account: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create account")
			return
		}

		// Convert to response format
		accountResponse := Account{
			ID:            account.ID,
			Number:        account.Number,
			Destination:   account.Destination,
			Description:   account.Description,
			AssociationID: account.AssociationID,
			IsActive:      account.IsActive,
			CreatedAt:     account.CreatedAt.Time,
			UpdatedAt:     account.UpdatedAt.Time,
		}

		RespondWithJSON(rw, http.StatusCreated, accountResponse)
	}
}

func HandleUpdateAccount(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		accountId, _ := strconv.Atoi(req.PathValue(AccountIdPathValue))

		// Parse request
		var accountRequest struct {
			Number      string `json:"number"`
			Destination string `json:"destination"`
			Description string `json:"description"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&accountRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate request
		if accountRequest.Number == "" {
			RespondWithError(rw, http.StatusBadRequest, "Account number is required")
			return
		}

		if accountRequest.Description == "" {
			RespondWithError(rw, http.StatusBadRequest, "Description is required")
			return
		}

		// Update account
		account, err := cfg.Db.UpdateAccount(req.Context(), database.UpdateAccountParams{
			ID:            int64(accountId),
			AssociationID: int64(associationId),
			Number:        accountRequest.Number,
			Destination:   accountRequest.Destination,
			Description:   accountRequest.Description,
		})

		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Account not found or doesn't belong to this association")
			} else {
				log.Printf("Error updating account: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Failed to update account")
			}
			return
		}

		// Convert to response format
		accountResponse := Account{
			ID:            account.ID,
			Number:        account.Number,
			Destination:   account.Destination,
			Description:   account.Description,
			AssociationID: account.AssociationID,
			IsActive:      account.IsActive,
			CreatedAt:     account.CreatedAt.Time,
			UpdatedAt:     account.UpdatedAt.Time,
		}

		RespondWithJSON(rw, http.StatusOK, accountResponse)
	}
}

func HandleDisableAccount(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		accountId, _ := strconv.Atoi(req.PathValue(AccountIdPathValue))

		// First get the account to verify it exists and belongs to the association
		account, err := cfg.Db.GetAccount(req.Context(), int64(accountId))
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Account not found")
			} else {
				log.Printf("Error getting account: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve account")
			}
			return
		}

		// Verify the account belongs to the association
		if account.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusForbidden, "Account does not belong to this association")
			return
		}

		// Check if the account is already inactive
		if !account.IsActive {
			RespondWithError(rw, http.StatusBadRequest, "Account is already inactive")
			return
		}

		// Disable the account
		err = cfg.Db.DisableAccount(req.Context(), database.DisableAccountParams{
			ID:            int64(accountId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			log.Printf("Error disabling account: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to disable account")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Account disabled successfully",
		})
	}
}
