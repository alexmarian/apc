package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func HandleUpdateUser(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	type parameters struct {
		Login    string `json:"login"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"isAdmin"`
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		request := parameters{}
		err := decoder.Decode(&request)
		if err != nil {
			var errors = fmt.Sprintf("Error decoding update user request: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error decoding update user request", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusBadRequest, errors)
			return
		}
		password, err := auth.HashPassword(request.Password)
		if err != nil {
			var errors = fmt.Sprintf("Error hashing password: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error hashing password", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		secret, qrCode, err := auth.GenerateQRCode(request.Login)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error generating QR code", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
			return
		}
		// Only an application admin may grant or revoke admin rights; anyone
		// else keeps their current flag whatever the body says.
		isAdmin := request.IsAdmin
		if claims := GetClaimsFromContext(req); claims == nil || !claims.IsAdmin {
			current, err := cfg.Db.GetUserByLogin(req.Context(), GetUserIdFromContext(req))
			if err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Error loading user")
				return
			}
			isAdmin = current.IsAdmin
		}
		user, err := cfg.Db.UpdateUserEmailAndPassword(req.Context(), database.UpdateUserEmailAndPasswordParams{
			PasswordHash: password,
			ToptSecret:   secret,
			IsAdmin:      isAdmin,
			Login:        GetUserIdFromContext(req),
		})
		if err != nil {
			var errors = fmt.Sprintf("Error creating user: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error creating user", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		type response struct {
			Login     string    `json:"login"`
			QrCode    string    `json:"qrCode"`
			CreatedAt time.Time `json:"createdAt"`
			UpdatedAt time.Time `json:"updatedAt"`
		}
		responseData := response{
			Login:     user.Login,
			QrCode:    qrCode,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}
		RespondWithJSON(rw, http.StatusOK, responseData)
	}
}

// UserRef is the public identity of a user, as shown in pickers and lists.
type UserRef struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
}
