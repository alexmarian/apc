package handlers

import (
	"fmt"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const AssociationIdPathValue = "associationId"

type Association struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Address       string    `json:"address"`
	Administrator string    `json:"administrator"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func HandleGetUserAssociations(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		userAssociationsIds := GetAssotiationIdsToContext(req)
		associationsFromList, err := cfg.Db.GetAssociationsFromList(req.Context(), userAssociationsIds)
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting associations", zap.String("userAssociationsIds", fmt.Sprintf("%v", userAssociationsIds)))
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		associations := make([]Association, len(associationsFromList))
		for i, association := range associationsFromList {
			associations[i] = Association{
				ID:            association.ID,
				Name:          association.Name,
				Address:       association.Address,
				Administrator: association.Administrator,
				CreatedAt:     association.CreatedAt.Time,
				UpdatedAt:     association.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusCreated, associations)
	}
}
func HandleGetUserAssociation(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		association, err := cfg.Db.GetAssociations(req.Context(), int64(associationId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting associations")
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}

		RespondWithJSON(rw, http.StatusCreated, &Association{
			ID:            association.ID,
			Name:          association.Name,
			Address:       association.Address,
			Administrator: association.Administrator,
			CreatedAt:     association.CreatedAt.Time,
			UpdatedAt:     association.UpdatedAt.Time,
		})
	}
}
