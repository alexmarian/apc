package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"time"
)

const RegistrationTokenPathValue = "token"

type RegistrationToken struct {
	Token       string     `json:"token"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   time.Time  `json:"expires_at"`
	UsedAt      *time.Time `json:"used_at,omitempty"`
	UsedBy      *string    `json:"used_by,omitempty"`
	RevokedAt   *time.Time `json:"revoked_at,omitempty"`
	RevokedBy   *string    `json:"revoked_by,omitempty"`
	Description string     `json:"description"`
	IsAdmin     bool       `json:"is_admin"`
	Status      string     `json:"status,omitempty"`
}

// Generate a secure random token
func generateToken() (string, error) {
	tokenBytes := make([]byte, 32)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(tokenBytes), nil
}

// Handler to create a new registration token (admin only)
func HandleCreateRegistrationToken(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Verify admin rights
		userLogin := GetUserIdFromContext(req)
		user, err := cfg.Db.GetUserByLogin(req.Context(), userLogin)
		if err != nil {
			log.Printf("Error getting user: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to verify user")
			return
		}

		if !user.IsAdmin {
			RespondWithError(rw, http.StatusForbidden, "Admin rights required")
			return
		}

		// Parse request
		var tokenRequest struct {
			ExpirationHours int    `json:"expiration_hours"`
			Description     string `json:"description"`
			IsAdmin         bool   `json:"is_admin"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&tokenRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate request
		if tokenRequest.ExpirationHours <= 0 {
			RespondWithError(rw, http.StatusBadRequest, "Expiration hours must be positive")
			return
		}

		if tokenRequest.Description == "" {
			RespondWithError(rw, http.StatusBadRequest, "Description is required")
			return
		}

		// Generate token
		token, err := generateToken()
		if err != nil {
			log.Printf("Error generating token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		// Calculate expiration time
		expiresAt := time.Now().Add(time.Duration(tokenRequest.ExpirationHours) * time.Hour)

		// Create token in database
		dbToken, err := cfg.Db.CreateRegistrationToken(req.Context(), database.CreateRegistrationTokenParams{
			Token:       token,
			CreatedBy:   userLogin,
			ExpiresAt:   expiresAt,
			Description: tokenRequest.Description,
			IsAdmin:     tokenRequest.IsAdmin,
		})

		if err != nil {
			log.Printf("Error creating token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create token")
			return
		}

		// Return created token
		tokenResponse := RegistrationToken{
			Token:       dbToken.Token,
			CreatedBy:   dbToken.CreatedBy,
			CreatedAt:   dbToken.CreatedAt,
			ExpiresAt:   dbToken.ExpiresAt,
			Description: dbToken.Description,
			IsAdmin:     dbToken.IsAdmin,
		}

		RespondWithJSON(rw, http.StatusCreated, tokenResponse)
	}
}

// Handler to revoke a registration token (admin only)
func HandleRevokeRegistrationToken(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Verify admin rights
		userLogin := GetUserIdFromContext(req)
		user, err := cfg.Db.GetUserByLogin(req.Context(), userLogin)
		if err != nil {
			log.Printf("Error getting user: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to verify user")
			return
		}

		if !user.IsAdmin {
			RespondWithError(rw, http.StatusForbidden, "Admin rights required")
			return
		}

		token := req.PathValue(RegistrationTokenPathValue)
		if token == "" {
			RespondWithError(rw, http.StatusBadRequest, "Token is required")
			return
		}

		// Revoke token
		err = cfg.Db.RevokeRegistrationToken(req.Context(), database.RevokeRegistrationTokenParams{
			RevokedBy: sql.NullString{String: userLogin, Valid: true},
			Token:     token,
		})

		if err != nil {
			log.Printf("Error revoking token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to revoke token")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Token revoked successfully",
		})
	}
}

// Handler to get all registration tokens with status (admin only)
func HandleGetAllRegistrationTokens(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Verify admin rights
		userLogin := GetUserIdFromContext(req)
		user, err := cfg.Db.GetUserByLogin(req.Context(), userLogin)
		if err != nil {
			log.Printf("Error getting user: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to verify user")
			return
		}

		if !user.IsAdmin {
			RespondWithError(rw, http.StatusForbidden, "Admin rights required")
			return
		}

		// Get all tokens with status
		dbTokens, err := cfg.Db.GetRegistrationTokensStatus(req.Context())
		if err != nil {
			log.Printf("Error getting tokens: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get tokens")
			return
		}

		// Convert to response format
		tokens := make([]RegistrationToken, len(dbTokens))
		for i, dbToken := range dbTokens {
			tokens[i] = RegistrationToken{
				Token:       dbToken.Token,
				CreatedBy:   dbToken.CreatedBy,
				CreatedAt:   dbToken.CreatedAt,
				ExpiresAt:   dbToken.ExpiresAt,
				Description: dbToken.Description,
				IsAdmin:     dbToken.IsAdmin,
				Status:      dbToken.Status,
			}

			if dbToken.UsedAt.Valid {
				usedAt := dbToken.UsedAt.Time
				tokens[i].UsedAt = &usedAt
			}

			if dbToken.UsedBy.Valid {
				usedBy := dbToken.UsedBy.String
				tokens[i].UsedBy = &usedBy
			}

			if dbToken.RevokedAt.Valid {
				revokedAt := dbToken.RevokedAt.Time
				tokens[i].RevokedAt = &revokedAt
			}

			if dbToken.RevokedBy.Valid {
				revokedBy := dbToken.RevokedBy.String
				tokens[i].RevokedBy = &revokedBy
			}
		}

		RespondWithJSON(rw, http.StatusOK, tokens)
	}
}
