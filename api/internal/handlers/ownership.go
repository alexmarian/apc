package handlers

import (
	"database/sql"
	"encoding/json"
	"github.com/alexmarian/apc/api/internal/database"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

func HandleDisableOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		// Parse path parameters
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		ownershipId, _ := strconv.Atoi(req.PathValue("ownershipId"))

		// Optional: Parse request body for additional data
		type DisableRequest struct {
			EndDate       *time.Time `json:"end_date"`
			DisableReason string     `json:"disable_reason"`
		}

		var disableReq DisableRequest
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&disableReq); err != nil && err != io.EOF {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Set default end date to current time if not provided
		endDate := time.Now()
		if disableReq.EndDate != nil {
			endDate = *disableReq.EndDate
		}

		// Get the ownership to validate it exists and belongs to the association
		ownership, err := cfg.Db.GetOwnership(req.Context(), int64(ownershipId))
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Ownership not found")
			} else {
				log.Printf("Error retrieving ownership: %s", err)
				RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve ownership")
			}
			return
		}

		// Validate association ID
		if ownership.AssociationID != int64(associationId) {
			RespondWithError(rw, http.StatusForbidden, "Ownership does not belong to this association")
			return
		}

		// Validate ownership is active
		if !ownership.IsActive {
			RespondWithError(rw, http.StatusBadRequest, "Ownership is already inactive")
			return
		}

		// Update ownership to inactive
		err = cfg.Db.DeactivateOwnership(req.Context(), database.DeactivateOwnershipParams{
			ID:      int64(ownershipId),
			EndDate: sql.NullTime{Time: endDate, Valid: true},
			// You might want to add a comments/notes field to ownerships table for storing disable_reason
		})

		if err != nil {
			log.Printf("Error deactivating ownership: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to deactivate ownership")
			return
		}

		// Return success response
		type DisableResponse struct {
			Message string    `json:"message"`
			EndDate time.Time `json:"end_date"`
		}

		response := DisableResponse{
			Message: "Ownership has been deactivated successfully",
			EndDate: endDate,
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}
