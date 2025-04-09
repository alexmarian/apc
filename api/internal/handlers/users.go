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

func HandleCreateUser(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		type request struct {
			Login    string `json:"login"`
			Password string `json:"password"`
			IsAdmin  bool   `json:"isAdmin"`
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
		password, err := auth.HashPassword(userData.Password)
		if err != nil {
			var errors = fmt.Sprintf("Error hashing password: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		secret, err := auth.GenerateTOTPSecret()
		if err != nil {
			log.Printf(err.Error())
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
		}
		qrCode, err := auth.GenerateQRCode(userData.Login, secret)
		if err != nil {
			log.Printf(err.Error())
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
		}
		user, err := cfg.Db.CreateUser(req.Context(), database.CreateUserParams{userData.Login, password, secret, userData.IsAdmin})
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
		}
		responseData := response{
			Login:     user.Login,
			QrCode:    qrCode,
			CreatedAt: user.CreatedAt,
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
		secret, err := auth.GenerateTOTPSecret()
		if err != nil {
			log.Printf(err.Error())
			RespondWithError(rw, http.StatusInternalServerError, err.Error())
		}
		qrCode, err := auth.GenerateQRCode(request.Login, secret)
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
