package handlers

import (
	"context"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
)

const userContextKey = "userID"
const assoctiationsContextKey = "associations"

type ApiConfig struct {
	Db     *database.Queries
	Secret string
}

type ErrorResponse struct {
	Msg  string `json:"msg"`
	Code int    `json:"code"`
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	data, err := json.Marshal(&ErrorResponse{msg, code})
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
func RespondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error encoding response: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func AddUserIdToContext(req *http.Request, userID string) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), userContextKey, userID))
}

func GetUserIdFromContext(req *http.Request) string {
	return req.Context().Value(userContextKey).(string)
}

func AddAssotiationIdsToContext(req *http.Request, associations []int64) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), assoctiationsContextKey, associations))
}

func GetAssotiationIdsToContext(req *http.Request) []int64 {
	return req.Context().Value(assoctiationsContextKey).([]int64)
}
