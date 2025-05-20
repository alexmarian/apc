package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
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
			logging.Logger.Log(zap.WarnLevel, "Error getting associations")
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
			logging.Logger.Log(zap.WarnLevel, "Error getting buildings")
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
			logging.Logger.Log(zap.WarnLevel, "Error getting buildings")
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
			logging.Logger.Log(zap.WarnLevel, "Error getting buildings")
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
				ID:                        owner.ID,
				UnitId:                    owner.UnitID,
				OwnerId:                   owner.OwnerID,
				OwnerName:                 owner.OwnerName,
				OwnerNormalizedName:       owner.OwnerNormalizedName,
				OwnerIdentificationNumber: owner.IdentificationNumber,
				AssociationId:             owner.AssociationID,
				StartDate:                 owner.StartDate.Time,
				EndDate:                   owner.EndDate.Time,
				IsActive:                  owner.IsActive,
				IsVoting:                  owner.IsVoting,
				RegistrationDocument:      owner.RegistrationDocument,
				RegistrationDate:          owner.RegistrationDate,
				CreatedAt:                 owner.CreatedAt.Time,
				UpdatedAt:                 owner.UpdatedAt.Time,
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

		applyUnitPartialUpdates(&updateRequest, existingUnit, &params)

		if err := cfg.Db.UpdateBuildingUnitById(req.Context(), params); err != nil {
			return
		}

		existingUnit.UpdatedAt = time.Now()

		RespondWithJSON(rw, http.StatusOK, existingUnit)
	}
}

func applyUnitPartialUpdates(update *UnitUpdateRequest, existing *Unit, params *database.UpdateBuildingUnitByIdParams) {
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

func HandleGetUnitReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		// First, get the unit details
		unit, err := cfg.Db.GetBuildingUnit(req.Context(), database.GetBuildingUnitParams{
			ID:         int64(unitId),
			BuildingID: int64(buildingId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit", zap.String("unitId", fmt.Sprintf("%d", unitId)))
			RespondWithError(rw, http.StatusNotFound, "Unit not found")
			return
		}

		// Get the building details
		building, err := cfg.Db.GetAssociationBuilding(req.Context(), database.GetAssociationBuildingParams{
			ID:            int64(buildingId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting building", zap.String("buildingId", fmt.Sprintf("%d", buildingId)))
			RespondWithError(rw, http.StatusNotFound, "Building not found")
			return
		}

		// Get all owners of this unit
		owners, err := cfg.Db.GetUnitOwners(req.Context(), database.GetUnitOwnersParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit owners", zap.String("unitId", fmt.Sprintf("%d", unitId)))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve unit owners")
			return
		}

		// Get ownership history
		ownerships, err := cfg.Db.GetUnitOwnerships(req.Context(), database.GetUnitOwnershipsParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit ownerships", zap.String("unitId", fmt.Sprintf("%d", unitId)))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve unit ownership history")
			return
		}

		// Prepare the response
		type UnitOwner struct {
			ID                   int64  `json:"id"`
			Name                 string `json:"name"`
			IdentificationNumber string `json:"identification_number"`
			ContactPhone         string `json:"contact_phone"`
			ContactEmail         string `json:"contact_email"`
			IsActive             bool   `json:"is_active"`
		}

		type OwnershipRecord struct {
			ID                   int64      `json:"id"`
			OwnerID              int64      `json:"owner_id"`
			OwnerName            string     `json:"owner_name"`
			StartDate            time.Time  `json:"start_date"`
			EndDate              *time.Time `json:"end_date,omitempty"`
			IsActive             bool       `json:"is_active"`
			RegistrationDocument string     `json:"registration_document"`
			RegistrationDate     time.Time  `json:"registration_date"`
		}

		type UnitReportResponse struct {
			UnitDetails struct {
				ID              int64     `json:"id"`
				CadastralNumber string    `json:"cadastral_number"`
				UnitNumber      string    `json:"unit_number"`
				Address         string    `json:"address"`
				Entrance        int64     `json:"entrance"`
				Area            float64   `json:"area"`
				Part            float64   `json:"part"`
				UnitType        string    `json:"unit_type"`
				Floor           int64     `json:"floor"`
				RoomCount       int64     `json:"room_count"`
				CreatedAt       time.Time `json:"created_at"`
				UpdatedAt       time.Time `json:"updated_at"`
			} `json:"unit_details"`

			BuildingDetails struct {
				ID              int64   `json:"id"`
				Name            string  `json:"name"`
				Address         string  `json:"address"`
				CadastralNumber string  `json:"cadastral_number"`
				TotalArea       float64 `json:"total_area"`
			} `json:"building_details"`

			CurrentOwners    []UnitOwner       `json:"current_owners"`
			OwnershipHistory []OwnershipRecord `json:"ownership_history"`
		}

		response := UnitReportResponse{}

		// Fill unit details
		response.UnitDetails.ID = unit.ID
		response.UnitDetails.CadastralNumber = unit.CadastralNumber
		response.UnitDetails.UnitNumber = unit.UnitNumber
		response.UnitDetails.Address = unit.Address
		response.UnitDetails.Entrance = unit.Entrance
		response.UnitDetails.Area = unit.Area
		response.UnitDetails.Part = unit.Part
		response.UnitDetails.UnitType = unit.UnitType
		response.UnitDetails.Floor = unit.Floor
		response.UnitDetails.RoomCount = unit.RoomCount
		response.UnitDetails.CreatedAt = unit.CreatedAt.Time
		response.UnitDetails.UpdatedAt = unit.UpdatedAt.Time

		// Fill building details
		response.BuildingDetails.ID = building.ID
		response.BuildingDetails.Name = building.Name
		response.BuildingDetails.Address = building.Address
		response.BuildingDetails.CadastralNumber = building.CadastralNumber
		response.BuildingDetails.TotalArea = building.TotalArea

		// Fill current owners
		currentOwners := []UnitOwner{}

		// Get active ownerships
		activeOwnerships, err := cfg.Db.GetActiveUnitOwnerships(req.Context(), database.GetActiveUnitOwnershipsParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting active ownerships", zap.String("unitId", fmt.Sprintf("%d", unitId)))
			// Continue anyway to return partial data
		}

		// Map active owner IDs
		activeOwnerIDs := make(map[int64]bool)
		for _, ownership := range activeOwnerships {
			activeOwnerIDs[ownership.OwnerID] = true
		}

		// Process owners
		for _, owner := range owners {
			currentOwners = append(currentOwners, UnitOwner{
				ID:                   owner.ID,
				Name:                 owner.Name,
				IdentificationNumber: owner.IdentificationNumber,
				ContactPhone:         owner.ContactPhone,
				ContactEmail:         owner.ContactEmail,
				IsActive:             activeOwnerIDs[owner.ID],
			})
		}

		response.CurrentOwners = currentOwners

		// Fill ownership history
		ownershipHistory := []OwnershipRecord{}

		// Create a map of owner IDs to names
		ownerNames := make(map[int64]string)
		for _, owner := range owners {
			ownerNames[owner.ID] = owner.Name
		}

		// Process ownerships
		for _, ownership := range ownerships {
			record := OwnershipRecord{
				ID:                   ownership.ID,
				OwnerID:              ownership.OwnerID,
				OwnerName:            ownerNames[ownership.OwnerID],
				StartDate:            ownership.StartDate.Time,
				IsActive:             ownership.IsActive,
				RegistrationDocument: ownership.RegistrationDocument,
				RegistrationDate:     ownership.RegistrationDate,
			}

			// Add end date if present
			if ownership.EndDate.Valid {
				endDate := ownership.EndDate.Time
				record.EndDate = &endDate
			}

			ownershipHistory = append(ownershipHistory, record)
		}

		response.OwnershipHistory = ownershipHistory

		RespondWithJSON(rw, http.StatusOK, response)
	}
}
