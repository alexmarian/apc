package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
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
		all, err := cfg.Db.ListAssociations(req.Context())
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error listing associations", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, fmt.Sprintf("Error getting associations: %s", err))
			return
		}
		result := make([]Association, len(all))
		for i, a := range all {
			result[i] = Association{
				ID:            a.ID,
				Name:          a.Name,
				Address:       a.Address,
				Administrator: a.Administrator,
				CreatedAt:     a.CreatedAt.Time,
				UpdatedAt:     a.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusOK, result)
	}
}

func HandleGetUserAssociation(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		association, err := cfg.Db.GetAssociations(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting association")
			RespondWithError(rw, http.StatusInternalServerError, fmt.Sprintf("Error getting association: %s", err))
			return
		}
		RespondWithJSON(rw, http.StatusOK, &Association{
			ID:            association.ID,
			Name:          association.Name,
			Address:       association.Address,
			Administrator: association.Administrator,
			CreatedAt:     association.CreatedAt.Time,
			UpdatedAt:     association.UpdatedAt.Time,
		})
	}
}
