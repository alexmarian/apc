package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func HandleLogin(cfg *ApiConfig) http.HandlerFunc {
	type parameters struct {
		Password         string `json:"password"`
		Login            string `json:"login"`
		TOTP             string `json:"totp"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}
	type response struct {
		Login        string `json:"login"`
		Token        string `json:"token,omitempty"`
		RefreshToken string `json:"refresh_token,omitempty"`
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		request := parameters{}
		if err := decoder.Decode(&request); err != nil {
			logging.Logger.Log(zapcore.WarnLevel, "Error decoding login request", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusBadRequest, fmt.Sprintf("Error decoding login request: %s", err))
			return
		}
		user, err := cfg.Db.GetUserByLogin(req.Context(), request.Login)
		if err != nil {
			logging.Logger.Log(zapcore.WarnLevel, "No user", zap.String("user", request.Login))
			RespondWithError(rw, http.StatusUnauthorized, "Login failure")
			return
		}
		if err := auth.CheckPasswordHash(request.Password, user.PasswordHash); err != nil {
			logging.Logger.Log(zapcore.WarnLevel, "Incorrect password", zap.String("user", request.Login))
			RespondWithError(rw, http.StatusUnauthorized, "Login failure")
			return
		}
		if ok, err := auth.VerifyTOTPCode(user.ToptSecret, request.TOTP); err != nil || !ok {
			logging.Logger.Log(zapcore.WarnLevel, "Incorrect totp", zap.String("user", request.Login))
			RespondWithError(rw, http.StatusUnauthorized, "Login failure")
			return
		}
		seconds := 3600
		if request.ExpiresInSeconds != 0 {
			seconds = request.ExpiresInSeconds
		}
		refreshToken, err := auth.MakeRefreshToken()
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating refresh token")
			return
		}
		if err := cfg.Db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
			Token: refreshToken,
			Login: user.Login,
		}); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating refresh token")
			return
		}
		token, _, err := auth.MakeJWT(user.Login, cfg.Secret, time.Duration(seconds)*time.Second, user.IsAdmin)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating token")
			return
		}
		RespondWithJSON(rw, http.StatusOK, response{
			Login:        user.Login,
			Token:        token,
			RefreshToken: refreshToken,
		})
	}
}

func HandleRefresh(cfg *ApiConfig) http.HandlerFunc {
	type response struct {
		Token string `json:"token,omitempty"`
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		refreshToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			logging.Logger.Log(zapcore.WarnLevel, "No refresh token")
			RespondWithError(rw, http.StatusUnauthorized, "Invalid token")
			return
		}
		rt, err := cfg.Db.GetValidRefreshToken(req.Context(), refreshToken)
		if err != nil {
			logging.Logger.Log(zapcore.WarnLevel, "No valid token", zap.String("token", refreshToken))
			RespondWithError(rw, http.StatusUnauthorized, "Invalid token")
			return
		}
		user, err := cfg.Db.GetUserByLogin(req.Context(), rt.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error fetching user")
			return
		}
		token, _, err := auth.MakeJWT(rt.Login, cfg.Secret, 3600*time.Second, user.IsAdmin)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating token")
			return
		}
		RespondWithJSON(rw, http.StatusOK, response{Token: token})
	}
}

func HandleLogout(cfg *ApiConfig) http.HandlerFunc {
	return func(rw http.ResponseWriter, req *http.Request) {
		claims := GetClaimsFromContext(req)
		if claims == nil {
			RespondWithError(rw, http.StatusUnauthorized, "unauthorized")
			return
		}
		expiresAt, _ := claims.GetExpirationTime()
		if err := cfg.Db.RevokeToken(req.Context(), claims.ID, expiresAt.Time); err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error revoking token")
			return
		}
		rw.WriteHeader(http.StatusNoContent)
	}
}
