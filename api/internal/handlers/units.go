package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"strconv"
	"time"
)

const UnitIdPathValue = "unitId"

type Unit struct {
	ID              int64     `json:"id"`
	CadastralNumber string    `json:"cadastral_number"`
	BuildingId      int64     `json:"building_id"`
	UnitNumber      string    `json:"unit_number"`
	Address         string    `json:"address"`
	Entrance        int64     `json:"entrance"`
	Area            float64   `json:"area"`
	Part            float64   `json:"part"`
	UnitType        string    `json:"unit_type"`
	Floor           int64     `json:"floor"`
	RoomCount       int64     `json:"room_count"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func HandleGetBuildingUnits(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		//associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitsFromList, err := cfg.Db.GetBuildingUnits(req.Context(), int64(buildingId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		units := make([]Unit, len(unitsFromList))
		for i, unit := range unitsFromList {
			units[i] = Unit{
				ID:              unit.ID,
				CadastralNumber: unit.CadastralNumber,
				BuildingId:      unit.BuildingID,
				UnitNumber:      unit.UnitNumber,
				Address:         unit.Address,
				Entrance:        unit.Entrance,
				Area:            unit.Area,
				Part:            unit.Part,
				UnitType:        unit.UnitType,
				Floor:           unit.Floor,
				RoomCount:       unit.RoomCount,
				CreatedAt:       unit.CreatedAt.Time,
				UpdatedAt:       unit.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusCreated, units)
	}
}
func HandleGetBuildingUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		//associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		unitPayload, err := getUnitPayloadFromDb(req.Context(), cfg, buildingId, unitId)
		if err != nil {
			var errors = fmt.Sprintf("Error getting buildings: %s", err)
			log.Printf(errors)
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Building not found")
				return
			}
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		RespondWithJSON(rw, http.StatusCreated, unitPayload)
	}
}

func HandleGetBuildingUnitOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		//associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		//buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		ownersFromDb, err := cfg.Db.GetUnitOwners(req.Context(), int64(unitId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting buildings: %s", err)
			log.Printf(errors)
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Building not found")
				return
			}
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		owners := make([]Owner, len(ownersFromDb))
		for i, owner := range ownersFromDb {
			owners[i] = Owner{
				ID:                   owner.ID,
				Name:                 owner.Name,
				NormalizedName:       owner.NormalizedName,
				IdentificationNumber: owner.IdentificationNumber,
				ContactPhone:         owner.ContactPhone,
				ContactEmail:         owner.ContactEmail,
				FirstDetectedAt:      owner.FirstDetectedAt.Time,
				CreatedAt:            owner.CreatedAt.Time,
				UpdatedAt:            owner.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusCreated, owners)
	}
}

func HandleUpdateBuildingUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		//associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		unit := Unit{}
		err := decoder.Decode(&unit)
		if err != nil {
			var errors = fmt.Sprintf("Error decoding update user request: %s", err)
			log.Printf(errors)
			RespondWithError(rw, http.StatusBadRequest, errors)
			return
		}
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		params := database.UpdateBuildingUnitByIdParams{
			ID:         int64(unitId),
			BuildingID: int64(buildingId),
		}

		unitPayload, err := getUnitPayloadFromDb(req.Context(), cfg, buildingId, unitId)
		applyPartialUpdates(unit, unitPayload, &params)

		err = cfg.Db.UpdateBuildingUnitById(req.Context(), params)
		if err != nil {
			var errors = fmt.Sprintf("Error getting buildings: %s", err)
			log.Printf(errors)
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Building not found")
				return
			}
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		if err != nil {
			var errors = fmt.Sprintf("Error getting buildings: %s", err)
			log.Printf(errors)
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Building not found")
				return
			}
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		RespondWithJSON(rw, http.StatusCreated, unitPayload)
	}
}

func applyPartialUpdates(unit Unit, existing *Unit, params *database.UpdateBuildingUnitByIdParams) {
	if unit.UnitNumber != "" {
		params.UnitNumber = unit.UnitNumber
		existing.UnitNumber = unit.UnitNumber
	} else {
		params.UnitNumber = existing.UnitNumber
	}
	if unit.Address != "" {
		params.Address = unit.Address
		existing.Address = unit.Address
	} else {
		params.Address = existing.Address
	}
	if unit.Entrance != 0 {
		params.Entrance = unit.Entrance
		existing.Entrance = unit.Entrance
	} else {
		params.Entrance = existing.Entrance
	}
	if unit.UnitType != "" {
		params.UnitType = unit.UnitType
		existing.UnitType = unit.UnitType
	} else {
		params.UnitType = existing.UnitType
	}
	if unit.Floor != 0 {
		params.Floor = unit.Floor
		existing.Floor = unit.Floor
	} else {
		params.Floor = existing.Floor
	}
	if unit.RoomCount != 0 {
		params.RoomCount = unit.RoomCount
		existing.RoomCount = unit.RoomCount
	} else {
		params.RoomCount = existing.RoomCount
	}
	existing.UpdatedAt = time.Now()
}

func getUnitPayloadFromDb(context context.Context, cfg *ApiConfig, buildingId int, unitId int) (*Unit, error) {
	unit, err := cfg.Db.GetBuildingUnit(context, database.GetBuildingUnitParams{
		BuildingID: int64(buildingId),
		ID:         int64(unitId),
	})
	if err != nil {
		return nil, err
	}

	unitPayload := &Unit{
		ID:              unit.ID,
		CadastralNumber: unit.CadastralNumber,
		BuildingId:      unit.BuildingID,
		UnitNumber:      unit.UnitNumber,
		Address:         unit.Address,
		Entrance:        unit.Entrance,
		Area:            unit.Area,
		Part:            unit.Part,
		UnitType:        unit.UnitType,
		Floor:           unit.Floor,
		RoomCount:       unit.RoomCount,
		CreatedAt:       unit.CreatedAt.Time,
		UpdatedAt:       unit.UpdatedAt.Time,
	}
	return unitPayload, nil
}
