package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"time"
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
		err := decoder.Decode(&request)
		if err != nil {
			var errors = fmt.Sprintf("Error decoding login user request: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusBadRequest, errors)
			return
		}
		user, err := cfg.Db.GetUserByLogin(req.Context(), request.Login)
		if err != nil {
			var errors = fmt.Sprintf("Error getting user: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		err = auth.CheckPasswordHash(request.Password, user.PasswordHash)
		if err != nil {
			RespondWithError(rw, http.StatusUnauthorized, "Incorrect email or password")
			return
		}
		if success, err := auth.VerifyTOTPCode(user.ToptSecret, request.TOTP); err != nil || !success {
			log.Printf("Invalid TOTP code")
			RespondWithError(rw, http.StatusUnauthorized, "Invalid TOTP code")
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
		err = cfg.Db.CreateRefreshToken(req.Context(), database.CreateRefreshTokenParams{
			Token: refreshToken,
			Login: user.Login,
		})
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating refresh token")
		}
		associations, err := cfg.Db.GetUserAssociationsByLogin(req.Context(), user.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error getting user associations")
			return
		}
		token, err := auth.MakeJWT(user.Login, cfg.Secret, time.Duration(seconds)*time.Second, associations)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating token")
			return
		}
		usr := response{
			Login:        user.Login,
			Token:        token,
			RefreshToken: refreshToken,
		}
		RespondWithJSON(rw, http.StatusOK, usr)
	}
}

func HandleRefresh(cfg *ApiConfig) http.HandlerFunc {
	type response struct {
		Token string `json:"token,omitempty"`
	}
	return func(rw http.ResponseWriter, req *http.Request) {
		refreshToken, err := auth.GetBearerToken(req.Header)
		if err != nil {
			RespondWithError(rw, http.StatusUnauthorized, "Invalid token")
			return
		}
		rt, err := cfg.Db.GetValidRefreshToken(req.Context(), refreshToken)
		if err != nil {
			RespondWithError(rw, http.StatusUnauthorized, "Invalid token")
			return
		}
		associations, err := cfg.Db.GetUserAssociationsByLogin(req.Context(), rt.Login)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error getting user associations")
			return
		}
		token, err := auth.MakeJWT(rt.Login, cfg.Secret, 3600*time.Second, associations)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Error creating token")
			return
		}
		resp := response{
			Token: token,
		}
		//rw.Header().Set("Set-Cookie", fmt.Sprintf("id=a3fWa; Expires=%s; Secure; HttpOnly", time.Now().Add(3600*time.Second).UTC().Format(time.RFC1123)))
		RespondWithJSON(rw, http.StatusOK, resp)
	}
}
