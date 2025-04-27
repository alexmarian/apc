package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/auth"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"time"
)

func HandleCreateUserWithToken(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Parse request
		type request struct {
			Login    string `json:"login"`
			Password string `json:"password"`
			Token    string `json:"token"` // Registration token
		}

		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		userData := request{}
		err := decoder.Decode(&userData)
		if err != nil {
			var errors = fmt.Sprintf("Error decoding create user request: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusBadRequest, errors)
			return
		}

		// Validate the token
		dbToken, err := cfg.Db.GetValidRegistrationToken(req.Context(), userData.Token)
		if err != nil {
			log.Printf("Error validating token: %s", err)
			RespondWithError(rw, http.StatusUnauthorized, "Invalid or expired registration token")
			return
		}

		// Hash the password
		hashedPassword, err := auth.HashPassword(userData.Password)
		if err != nil {
			var errors = fmt.Sprintf("Error hashing password: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}

		// Generate TOTP secret
		secret, qrCode, err := auth.GenerateQRCode(userData.Login)
		if err != nil {
			log.Printf(err.Error())
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
			return
		}

		// Create the user
		user, err := cfg.Db.CreateUser(req.Context(), database.CreateUserParams{
			Login:        userData.Login,
			PasswordHash: hashedPassword,
			ToptSecret:   secret,
			IsAdmin:      dbToken.IsAdmin, // Use isAdmin from token
		})
		if err != nil {
			var errors = fmt.Sprintf("Error creating user: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}

		// Mark token as used
		err = cfg.Db.UseRegistrationToken(req.Context(), database.UseRegistrationTokenParams{
			UsedBy: sql.NullString{String: userData.Login, Valid: true},
			Token:  userData.Token,
		})
		if err != nil {
			log.Printf("Error marking token as used: %s", err)
			// Don't fail the request if this happens
		}

		// Return response
		type response struct {
			Login     string    `json:"login"`
			QrCode    string    `json:"qrCode"`
			CreatedAt time.Time `json:"createdAt"`
			IsAdmin   bool      `json:"isAdmin"`
		}
		responseData := response{
			Login:     user.Login,
			QrCode:    qrCode,
			CreatedAt: user.CreatedAt,
			IsAdmin:   user.IsAdmin,
		}
		RespondWithJSON(rw, http.StatusCreated, responseData)
	}
}

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
			log.Printf(errors)
			RespondWithError(rw, http.StatusBadRequest, errors)
			return
		}
		password, err := auth.HashPassword(request.Password)
		if err != nil {
			var errors = fmt.Sprintf("Error hashing password: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		secret, qrCode, err := auth.GenerateQRCode(request.Login)
		if err != nil {
			log.Printf(err.Error())
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
		}
		user, err := cfg.Db.UpdateUserEmailAndPassword(req.Context(), database.UpdateUserEmailAndPasswordParams{
			password, secret, request.IsAdmin, GetUserIdFromContext(req)})
		if err != nil {
			var errors = fmt.Sprintf("Error creating user: %s", err)
			log.Printf(errors)
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
