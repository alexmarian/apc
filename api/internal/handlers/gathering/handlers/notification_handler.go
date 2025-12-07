package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// NotificationHandler handles notification operations
type NotificationHandler struct {
	cfg *handlers.ApiConfig
}

// NewNotificationHandler creates a new NotificationHandler
func NewNotificationHandler(cfg *handlers.ApiConfig) *NotificationHandler {
	return &NotificationHandler{
		cfg: cfg,
	}
}

// HandleSendNotification sends notifications to owners
func (h *NotificationHandler) HandleSendNotification() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		var notifyReq struct {
			NotificationType string  `json:"notification_type"` // invitation, reminder, results
			OwnerIDs         []int64 `json:"owner_ids,omitempty"`
			SendVia          string  `json:"send_via"` // email, sms
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&notifyReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// If no specific owners, could get non-participating owners here
		// (implementation depends on specific notification requirements)

		// Send notifications
		var sent []database.VotingNotification
		for _, ownerID := range notifyReq.OwnerIDs {
			notification, err := h.cfg.Db.CreateNotification(req.Context(), database.CreateNotificationParams{
				GatheringID:      int64(gatheringID),
				OwnerID:          ownerID,
				NotificationType: notifyReq.NotificationType,
				SentVia:          sql.NullString{String: notifyReq.SendVia, Valid: true},
			})

			if err == nil {
				sent = append(sent, notification)
				// Here you would actually send the notification via email/SMS
			}
		}

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"sent_count":    len(sent),
			"notifications": sent,
		})
	}
}

// HandleGetAuditLogs returns audit logs for a gathering
func (h *NotificationHandler) HandleGetAuditLogs() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		limitStr := req.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		logs, err := h.cfg.Db.GetAuditLogs(req.Context(), database.GetAuditLogsParams{
			GatheringID: int64(gatheringID),
			Limit:       int64(limit),
		})

		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get audit logs")
			return
		}

		handlers.RespondWithJSON(rw, http.StatusOK, logs)
	}
}
