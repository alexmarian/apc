package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"time"
)

const userContextKey = "userID"
const assoctiationsContextKey = "associations"
const startDateQueryKey = "start_date"
const endDateQueryKey = "end_date"

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

func GetRequestDateRange(req *http.Request, rw *http.ResponseWriter) (startDate, endDate time.Time, err error) {
	startDateStr := req.URL.Query().Get(startDateQueryKey)
	endDateStr := req.URL.Query().Get(endDateQueryKey)

	if startDateStr == "" {
		startDate = time.Now().AddDate(0, -1, 0)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			err = fmt.Errorf("invalid start_date format: %w", err)
			return
		}
	}
	if endDateStr == "" {
		endDate = time.Now()
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			err = fmt.Errorf("invalid end_date format: %w", err)
			return
		}
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	}
	return
}
