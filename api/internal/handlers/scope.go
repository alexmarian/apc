package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// MiddlewareAssociationResource only checks the caller may access the
// association in the path. These resolve the rest of the path *within* that
// association, so an association admin can't reach another association's
// buildings or units by pairing their own association id with foreign ids.
// On failure they have already written the response (404 when the resource
// isn't in the association).

func pathInt64(req *http.Request, name string) (int64, bool) {
	v, err := strconv.ParseInt(req.PathValue(name), 10, 64)
	return v, err == nil && v > 0
}

// associationBuilding resolves {associationId}/{buildingId}.
func associationBuilding(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.Building, bool) {
	associationId, okA := pathInt64(req, AssociationIdPathValue)
	buildingId, okB := pathInt64(req, BuildingIdPathValue)
	if !okA || !okB {
		RespondWithError(rw, http.StatusBadRequest, "Invalid id")
		return database.Building{}, false
	}
	building, err := cfg.Db.GetAssociationBuilding(req.Context(), database.GetAssociationBuildingParams{
		ID:            buildingId,
		AssociationID: associationId,
	})
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(rw, http.StatusNotFound, "Building not found")
		return database.Building{}, false
	}
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting building", zap.String("error", err.Error()))
		RespondWithError(rw, http.StatusInternalServerError, "Error getting building")
		return database.Building{}, false
	}
	return building, true
}

// associationUnit resolves {associationId}/{buildingId}/{unitId}.
func associationUnit(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.Building, database.Unit, bool) {
	building, ok := associationBuilding(cfg, rw, req)
	if !ok {
		return building, database.Unit{}, false
	}
	unitId, ok := pathInt64(req, UnitIdPathValue)
	if !ok {
		RespondWithError(rw, http.StatusBadRequest, "Invalid unit id")
		return building, database.Unit{}, false
	}
	unit, err := cfg.Db.GetBuildingUnit(req.Context(), database.GetBuildingUnitParams{ID: unitId, BuildingID: building.ID})
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(rw, http.StatusNotFound, "Unit not found")
		return building, database.Unit{}, false
	}
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting unit", zap.String("error", err.Error()))
		RespondWithError(rw, http.StatusInternalServerError, "Error getting unit")
		return building, database.Unit{}, false
	}
	return building, unit, true
}
