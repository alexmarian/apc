package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
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
	// Associations the invited user is linked to on registration (never
	// set for admin invitations).
	Associations []AssociationRef `json:"associations"`
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

		// Parse request
		var tokenRequest struct {
			ExpirationHours int     `json:"expiration_hours"`
			Description     string  `json:"description"`
			IsAdmin         bool    `json:"is_admin"`
			AssociationIds  []int64 `json:"association_ids"`
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

		if tokenRequest.IsAdmin && len(tokenRequest.AssociationIds) > 0 {
			RespondWithError(rw, http.StatusBadRequest, "Admin invitations can't be limited to associations")
			return
		}
		associationIds, ok := validAssociationIds(cfg, rw, req, tokenRequest.AssociationIds)
		if !ok {
			return
		}

		// Generate token
		token, err := generateToken()
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error generating token", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		// Calculate expiration time
		expiresAt := time.Now().UTC().Add(time.Duration(tokenRequest.ExpirationHours) * time.Hour)

		// Create the token and its association links together.
		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create token")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)
		dbToken, err := q.CreateRegistrationToken(ctx, database.CreateRegistrationTokenParams{
			Token:       token,
			CreatedBy:   userLogin,
			ExpiresAt:   expiresAt,
			Description: tokenRequest.Description,
			IsAdmin:     tokenRequest.IsAdmin,
		})
		for _, id := range associationIds {
			if err != nil {
				break
			}
			err = q.AddRegistrationTokenAssociation(ctx, database.AddRegistrationTokenAssociationParams{Token: token, AssociationID: id})
		}
		if err == nil {
			err = tx.Commit()
		}

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating token", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create token")
			return
		}

		// Return created token
		tokenResponse := RegistrationToken{
			Token:        dbToken.Token,
			CreatedBy:    dbToken.CreatedBy,
			CreatedAt:    dbToken.CreatedAt,
			ExpiresAt:    dbToken.ExpiresAt,
			Description:  dbToken.Description,
			IsAdmin:      dbToken.IsAdmin,
			Status:       "Valid",
			Associations: tokenAssociations(req, cfg.Db)[dbToken.Token],
		}
		if tokenResponse.Associations == nil {
			tokenResponse.Associations = []AssociationRef{}
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
			logging.Logger.Log(zap.WarnLevel, "Error getting user", zap.String("userLogin", userLogin))
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
			logging.Logger.Log(zap.WarnLevel, "Error revoking token", zap.String("token", token), zap.String("error", err.Error()))
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
			logging.Logger.Log(zap.WarnLevel, "Error getting user", zap.String("userLogin", userLogin))
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
			logging.Logger.Log(zap.WarnLevel, "Error getting tokens", zap.String("userLogin", userLogin))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get tokens")
			return
		}

		// Convert to response format
		associations := tokenAssociations(req, cfg.Db)
		tokens := make([]RegistrationToken, len(dbTokens))
		for i, dbToken := range dbTokens {
			tokens[i] = RegistrationToken{
				Token:        dbToken.Token,
				CreatedBy:    dbToken.CreatedBy,
				CreatedAt:    dbToken.CreatedAt,
				ExpiresAt:    dbToken.ExpiresAt,
				Description:  dbToken.Description,
				IsAdmin:      dbToken.IsAdmin,
				Status:       dbToken.Status,
				Associations: associations[dbToken.Token],
			}
			if tokens[i].Associations == nil {
				tokens[i].Associations = []AssociationRef{}
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

// tokenAssociations groups every registration token's associations by
// token. Errors yield an empty map (associations are informational here).
func tokenAssociations(req *http.Request, q *database.Queries) map[string][]AssociationRef {
	byToken := make(map[string][]AssociationRef)
	rows, err := q.ListRegistrationTokenAssociations(req.Context())
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error listing token associations", zap.String("error", err.Error()))
		return byToken
	}
	for _, r := range rows {
		byToken[r.Token] = append(byToken[r.Token], AssociationRef{ID: r.ID, Name: r.Name})
	}
	return byToken
}
