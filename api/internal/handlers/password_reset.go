package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
)

// HandleRequestPasswordReset (admin-only) creates a one-hour password
// reset token for a user and returns it, for the admin to send as a
// /reset?token=… link. Unknown login → 404.
func HandleRequestPasswordReset(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Login string `json:"login"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if _, err := cfg.Db.GetUserByLogin(req.Context(), request.Login); err != nil {
			RespondWithError(rw, http.StatusNotFound, "User not found")
			return
		}

		token := make([]byte, 32)
		if _, err := rand.Read(token); err != nil {
			log.Printf("Error generating reset token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error generating reset token")
			return
		}
		resetToken := hex.EncodeToString(token)
		expiresAt := time.Now().UTC().Add(1 * time.Hour)
		if err := cfg.Db.CreatePasswordResetToken(req.Context(), database.CreatePasswordResetTokenParams{
			Token:     resetToken,
			Login:     request.Login,
			ExpiresAt: expiresAt,
		}); err != nil {
			log.Printf("Error storing reset token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]any{
			"message":   "Password reset token has been generated",
			"token":     resetToken,
			"expiresAt": expiresAt,
		})
	}
}

// HandleResetPassword sets a new password with a reset token. With
// reset_totp_secret it also issues a new TOTP secret and makes the user
// pending again, so they confirm it through onboarding before logging in.
// Admin status is never touched.
func HandleResetPassword(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Token           string `json:"token"`
			NewPassword     string `json:"new_password"`
			ResetTOTPSecret bool   `json:"reset_totp_secret"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		if msg := validatePassword(request.NewPassword); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		passwordHash, err := auth.HashPassword(request.NewPassword)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}

		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)

		resetToken, err := q.GetValidPasswordResetToken(ctx, request.Token)
		if errors.Is(err, sql.ErrNoRows) {
			RespondWithError(rw, http.StatusBadRequest, "Invalid or expired reset token")
			return
		}
		if err != nil {
			log.Printf("Error validating reset token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}
		user, err := q.GetUserByLogin(ctx, resetToken.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}
		if err := q.SetUserPassword(ctx, database.SetUserPasswordParams{PasswordHash: passwordHash, ID: user.ID}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error resetting password")
			return
		}
		var enrollment *auth.TOTPEnrollment
		if request.ResetTOTPSecret {
			e, err := auth.GenerateTOTPEnrollment(user.Login)
			if err == nil {
				err = q.SetUserTotpSecret(ctx, database.SetUserTotpSecretParams{ToptSecret: e.Secret, ID: user.ID})
			}
			if err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Error resetting two-factor authentication")
				return
			}
			enrollment = &e
		}
		if err := q.UsePasswordResetToken(ctx, request.Token); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}
		if err := tx.Commit(); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}

		response := map[string]any{
			"message":            "Password has been reset successfully",
			"login":              user.Login,
			"onboardingRequired": enrollment != nil || !user.TotpConfirmedAt.Valid,
		}
		if enrollment != nil {
			response["enrollment"] = toEnrollmentResponse(*enrollment)
			response["qrCode"] = enrollment.QRCode // legacy Vue reset page
		}
		RespondWithJSON(rw, http.StatusOK, response)
	}
}
