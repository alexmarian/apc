package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"time"
)

func HandleRequestPasswordReset(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Login string `json:"login"`
		}

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		if err := decoder.Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		_, err := cfg.Db.GetUserByLogin(req.Context(), request.Login)
		if err != nil {
			RespondWithJSON(rw, http.StatusOK, map[string]string{
				"message": "If this account exists, a password reset token has been generated",
			})
			return
		}

		token := make([]byte, 32)
		_, err = rand.Read(token)
		if err != nil {
			log.Printf("Error generating reset token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error generating reset token")
			return
		}
		resetToken := hex.EncodeToString(token)

		// Store the token with expiration (24 hours)
		expiresAt := time.Now().Add(1 * time.Hour)
		err = cfg.Db.CreatePasswordResetToken(req.Context(), database.CreatePasswordResetTokenParams{
			Token:     resetToken,
			Login:     request.Login,
			ExpiresAt: expiresAt,
		})

		if err != nil {
			log.Printf("Error storing reset token: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{
			"message": "Password reset token has been generated",
			"token":   resetToken, // In production, remove this and send via email instead
		})
	}
}

func HandleResetPassword(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		var request struct {
			Token           string `json:"token"`
			NewPassword     string `json:"new_password"`
			ResetTOTPSecret bool   `json:"reset_totp_secret"` // Optional: reset TOTP seed
		}

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		if err := decoder.Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate token
		resetToken, err := cfg.Db.GetValidPasswordResetToken(req.Context(), request.Token)
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusBadRequest, "Invalid or expired reset token")
			} else {
				log.Printf("Error validating reset token: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			}
			return
		}

		passwordHash, err := auth.HashPassword(request.NewPassword)
		if err != nil {
			log.Printf("Error hashing password: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
			return
		}

		// Generate new TOTP secret if requested
		var totpSecret string
		var qrCode string

		if request.ResetTOTPSecret {
			totpSecret, qrCode, err = auth.GenerateQRCode(resetToken.Login)
			if err != nil {
				log.Printf("Error generating TOTP: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
				return
			}
		} else {
			// Get the existing TOTP secret
			user, err := cfg.Db.GetUserByLogin(req.Context(), resetToken.Login)
			if err != nil {
				log.Printf("Error fetching user: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Error processing reset request")
				return
			}
			totpSecret = user.ToptSecret
		}

		// Update user password and possibly TOTP secret
		_, err = cfg.Db.UpdateUserEmailAndPassword(req.Context(), database.UpdateUserEmailAndPasswordParams{
			PasswordHash: passwordHash,
			ToptSecret:   totpSecret,
			IsAdmin:      false, // Don't change admin status during reset
			Login:        resetToken.Login,
		})

		if err != nil {
			log.Printf("Error updating user: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Error resetting password")
			return
		}

		err = cfg.Db.UsePasswordResetToken(req.Context(), request.Token)
		if err != nil {
			log.Printf("Error marking token as used: %s", err)
		}

		response := map[string]string{
			"message": "Password has been reset successfully",
		}

		if request.ResetTOTPSecret {
			response["qrCode"] = qrCode
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}
