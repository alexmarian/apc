package handlers

import (
	"database/sql"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

const BuildingIdPathValue = "buildingId"

type Building struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Address         string    `json:"address"`
	CadastralNumber string    `json:"cadastral_number"`
	TotalArea       float64   `json:"total_area"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func HandleGetAssociationBuildings(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		buildingsFromList, err := cfg.Db.GetAssociationBuildings(req.Context(), int64(associationId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting associations")
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		buildings := make([]Building, len(buildingsFromList))
		for i, building := range buildingsFromList {
			buildings[i] = Building{
				ID:              building.ID,
				Name:            building.Name,
				Address:         building.Address,
				CadastralNumber: building.CadastralNumber,
				TotalArea:       building.TotalArea,
				CreatedAt:       building.CreatedAt.Time,
				UpdatedAt:       building.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusCreated, buildings)
	}
}
func HandleGetAssociationBuilding(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))

		building, err := cfg.Db.GetAssociationBuilding(req.Context(), database.GetAssociationBuildingParams{
			AssociationID: int64(associationId),
			ID:            int64(buildingId),
		})
		if err != nil {
			var errors = fmt.Sprintf("Error getting buildings: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting associations")
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Building not found")
				return
			}
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}

		RespondWithJSON(rw, http.StatusCreated, &Building{
			ID:              building.ID,
			Name:            building.Name,
			Address:         building.Address,
			CadastralNumber: building.CadastralNumber,
			TotalArea:       building.TotalArea,
			CreatedAt:       building.CreatedAt.Time,
			UpdatedAt:       building.UpdatedAt.Time,
		})
	}
}
