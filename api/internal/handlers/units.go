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

type UnitUpdateRequest struct {
	UnitNumber *string  `json:"unit_number,omitempty"`
	Address    *string  `json:"address,omitempty"`
	Entrance   *int64   `json:"entrance,omitempty"`
	Area       *float64 `json:"area,omitempty"`
	Part       *float64 `json:"part,omitempty"`
	UnitType   *string  `json:"unit_type,omitempty"`
	Floor      *int64   `json:"floor,omitempty"`
	RoomCount  *int64   `json:"room_count,omitempty"`
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
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		//buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		ownersFromDb, err := cfg.Db.GetUnitOwners(req.Context(), database.GetUnitOwnersParams{
			AssociationID: int64(associationId),
			UnitID:        int64(unitId),
		})
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

func HandleGetBuildingUnitOwnerships(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		//buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		ownershipsFromDb, err := cfg.Db.GetUnitOwnerships(req.Context(), database.GetUnitOwnershipsParams{
			AssociationID: int64(associationId),
			UnitID:        int64(unitId),
		})
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
		ownerships := make([]Ownership, len(ownershipsFromDb))
		for i, owner := range ownershipsFromDb {
			ownerships[i] = Ownership{
				ID:                   owner.ID,
				UnitId:               owner.UnitID,
				OwnerId:              owner.OwnerID,
				AssociationId:        owner.AssociationID,
				StartDate:            owner.StartDate.Time,
				EndDate:              owner.EndDate.Time,
				IsActive:             owner.IsActive,
				RegistrationDocument: owner.RegistrationDocument,
				RegistrationDate:     owner.RegistrationDate,
				CreatedAt:            owner.CreatedAt.Time,
				UpdatedAt:            owner.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusOK, ownerships)
	}
}

func HandleUpdateBuildingUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		existingUnit, err := getUnitPayloadFromDb(req.Context(), cfg, buildingId, unitId)
		if err != nil {
			return
		}

		decoder := json.NewDecoder(req.Body)
		var updateRequest UnitUpdateRequest
		if err := decoder.Decode(&updateRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		params := database.UpdateBuildingUnitByIdParams{
			ID:         int64(unitId),
			BuildingID: int64(buildingId),
		}

		applyPartialUpdates(&updateRequest, existingUnit, &params)

		if err := cfg.Db.UpdateBuildingUnitById(req.Context(), params); err != nil {
			return
		}

		existingUnit.UpdatedAt = time.Now()

		RespondWithJSON(rw, http.StatusOK, existingUnit)
	}
}

func applyPartialUpdates(update *UnitUpdateRequest, existing *Unit, params *database.UpdateBuildingUnitByIdParams) {
	// Use pointers to distinguish between "field not provided" and "field set to zero value"
	if update.UnitNumber != nil {
		params.UnitNumber = *update.UnitNumber
		existing.UnitNumber = *update.UnitNumber
	} else {
		params.UnitNumber = existing.UnitNumber
	}

	if update.Address != nil {
		params.Address = *update.Address
		existing.Address = *update.Address
	} else {
		params.Address = existing.Address
	}

	if update.Entrance != nil {
		params.Entrance = *update.Entrance
		existing.Entrance = *update.Entrance
	} else {
		params.Entrance = existing.Entrance
	}

	if update.UnitType != nil {
		params.UnitType = *update.UnitType
		existing.UnitType = *update.UnitType
	} else {
		params.UnitType = existing.UnitType
	}

	if update.Floor != nil {
		params.Floor = *update.Floor
		existing.Floor = *update.Floor
	} else {
		params.Floor = existing.Floor
	}

	if update.RoomCount != nil {
		params.RoomCount = *update.RoomCount
		existing.RoomCount = *update.RoomCount
	} else {
		params.RoomCount = existing.RoomCount
	}

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
