package handlers

import (
	"context"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
)

const userContextKey = "userID"

type ApiConfig struct {
	Db     *database.Queries
	Secret string
}

func RespondWithError(w http.ResponseWriter, code int, msg string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(msg))
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
