package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
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
	OwnerNames      string    `json:"owner_names,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// UnitUpdateRequest is a partial update: absent fields keep their value.
// Area, Part and CadastralNumber are app-admin only.
type UnitUpdateRequest struct {
	CadastralNumber *string  `json:"cadastral_number,omitempty"`
	UnitNumber      *string  `json:"unit_number,omitempty"`
	Address         *string  `json:"address,omitempty"`
	Entrance        *int64   `json:"entrance,omitempty"`
	Area            *float64 `json:"area,omitempty"`
	Part            *float64 `json:"part,omitempty"`
	UnitType        *string  `json:"unit_type,omitempty"`
	Floor           *int64   `json:"floor,omitempty"`
	RoomCount       *int64   `json:"room_count,omitempty"`
}

func toUnit(u database.Unit) Unit {
	return Unit{
		ID:              u.ID,
		CadastralNumber: u.CadastralNumber,
		BuildingId:      u.BuildingID,
		UnitNumber:      u.UnitNumber,
		Address:         u.Address,
		Entrance:        u.Entrance,
		Area:            u.Area,
		Part:            u.Part,
		UnitType:        u.UnitType,
		Floor:           u.Floor,
		RoomCount:       u.RoomCount,
		CreatedAt:       u.CreatedAt.Time,
		UpdatedAt:       u.UpdatedAt.Time,
	}
}

func HandleGetBuildingUnits(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, ok := associationBuilding(cfg, rw, req)
		if !ok {
			return
		}
		unitsFromList, err := cfg.Db.GetBuildingUnitsWithOwners(req.Context(), database.GetBuildingUnitsWithOwnersParams{
			AssociationID: building.AssociationID,
			BuildingID:    building.ID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting units", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error getting units")
			return
		}
		units := make([]Unit, len(unitsFromList))
		for i, unit := range unitsFromList {
			ownerNames := ""
			if unit.OwnerNames != nil {
				ownerNames = fmt.Sprintf("%v", unit.OwnerNames)
			}
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
				OwnerNames:      ownerNames,
				CreatedAt:       unit.CreatedAt.Time,
				UpdatedAt:       unit.UpdatedAt.Time,
			}
		}
		RespondWithJSON(rw, http.StatusOK, units)
	}
}

func HandleGetBuildingUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		_, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		RespondWithJSON(rw, http.StatusOK, toUnit(unit))
	}
}

func HandleGetBuildingUnitOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		ownersFromDb, err := cfg.Db.GetUnitOwners(req.Context(), database.GetUnitOwnersParams{
			AssociationID: building.AssociationID,
			UnitID:        unit.ID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit owners", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error getting unit owners")
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
		RespondWithJSON(rw, http.StatusOK, owners)
	}
}

func HandleGetBuildingUnitOwnerships(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		ownershipsFromDb, err := cfg.Db.GetUnitOwnerships(req.Context(), database.GetUnitOwnershipsParams{
			AssociationID: building.AssociationID,
			UnitID:        unit.ID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ownerships", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error getting ownerships")
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

// Unit fields an update may set. Area, part and cadastral number are
// read-only: they come from the cadastral records units are imported from.
const (
	unitEntranceMin  = -2
	unitTextMaxRunes = 255
)

// validateUnitUpdate returns a message for a 400, or "".
func validateUnitUpdate(u *UnitUpdateRequest, isAdmin bool) string {
	trim := func(p *string) {
		if p != nil {
			*p = strings.TrimSpace(*p)
		}
	}
	trim(u.UnitNumber)
	trim(u.Address)
	trim(u.UnitType)
	trim(u.CadastralNumber)
	structural := u.Area != nil || u.Part != nil || u.CadastralNumber != nil
	switch {
	case structural && !isAdmin:
		return "Area, share and cadastral number can only be changed by an application admin"
	case u.CadastralNumber != nil && *u.CadastralNumber == "":
		return "Cadastral number is required"
	case u.Area != nil && *u.Area <= 0:
		return "Area must be greater than 0"
	case u.Part != nil && (*u.Part <= 0 || *u.Part > 1):
		return "Share must be greater than 0 and at most 100%"

	case u.UnitNumber != nil && *u.UnitNumber == "":
		return "Unit number is required"
	case u.UnitNumber != nil && len([]rune(*u.UnitNumber)) > unitTextMaxRunes,
		u.Address != nil && len([]rune(*u.Address)) > unitTextMaxRunes:
		return "Text fields must be at most 255 characters"
	case u.UnitType != nil && *u.UnitType == "":
		return "Unit type is required"
	case u.Entrance != nil && *u.Entrance < unitEntranceMin:
		return "Entrance must be at least -2"
	case u.RoomCount != nil && *u.RoomCount < 0:
		return "Room count must be at least 0"
	}
	return ""
}

// HandleUpdateBuildingUnit partially updates a unit (only the fields
// present in the body change). Area, share and cadastral number are
// app-admin only (they come from cadastral records).
func HandleUpdateBuildingUnit(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		var update UnitUpdateRequest
		if err := json.NewDecoder(req.Body).Decode(&update); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		claims := GetClaimsFromContext(req)
		if msg := validateUnitUpdate(&update, claims != nil && claims.IsAdmin); msg != "" {
			RespondWithError(rw, http.StatusBadRequest, msg)
			return
		}
		if update.CadastralNumber != nil && *update.CadastralNumber != unit.CadastralNumber {
			if _, err := cfg.Db.GetUnitByCadastralNumber(req.Context(), *update.CadastralNumber); err == nil {
				RespondWithError(rw, http.StatusConflict, "Cadastral number already used by another unit")
				return
			}
		}

		pick := func(p *string, v string) string {
			if p != nil {
				return *p
			}
			return v
		}
		pickInt := func(p *int64, v int64) int64 {
			if p != nil {
				return *p
			}
			return v
		}
		pickFloat := func(p *float64, v float64) float64 {
			if p != nil {
				return *p
			}
			return v
		}
		params := database.UpdateUnitAllParams{
			CadastralNumber: pick(update.CadastralNumber, unit.CadastralNumber),
			UnitNumber:      pick(update.UnitNumber, unit.UnitNumber),
			Address:         pick(update.Address, unit.Address),
			Entrance:        pickInt(update.Entrance, unit.Entrance),
			Area:            pickFloat(update.Area, unit.Area),
			Part:            pickFloat(update.Part, unit.Part),
			UnitType:        pick(update.UnitType, unit.UnitType),
			Floor:           pickInt(update.Floor, unit.Floor),
			RoomCount:       pickInt(update.RoomCount, unit.RoomCount),
			ID:              unit.ID,
			BuildingID:      building.ID,
		}
		if err := cfg.Db.UpdateUnitAll(req.Context(), params); err != nil {
			if isUniqueViolation(err) {
				RespondWithError(rw, http.StatusConflict, "Cadastral number already used by another unit")
				return
			}
			logging.Logger.Log(zap.WarnLevel, "Error updating unit", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update unit")
			return
		}
		updated := database.Unit{
			ID: unit.ID, BuildingID: building.ID, CadastralNumber: params.CadastralNumber, UnitNumber: params.UnitNumber,
			Address: params.Address, Entrance: params.Entrance, Area: params.Area, Part: params.Part,
			UnitType: params.UnitType, Floor: params.Floor, RoomCount: params.RoomCount, CreatedAt: unit.CreatedAt,
		}
		result := toUnit(updated)
		result.UpdatedAt = time.Now()
		RespondWithJSON(rw, http.StatusOK, result)
	}
}

// HandleGetUnitReport returns a unit with its building, owners (current and
// past, flagged) and full ownership history — everything the unit detail
// page shows, in one call.
func HandleGetUnitReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
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
		IsVoting             bool       `json:"is_voting"`
		RegistrationDocument string     `json:"registration_document"`
		RegistrationDate     time.Time  `json:"registration_date"`
	}
	type UnitDetails struct {
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
	}
	type BuildingDetails struct {
		ID              int64   `json:"id"`
		Name            string  `json:"name"`
		Address         string  `json:"address"`
		CadastralNumber string  `json:"cadastral_number"`
		TotalArea       float64 `json:"total_area"`
	}
	type UnitReportResponse struct {
		UnitDetails      UnitDetails       `json:"unit_details"`
		BuildingDetails  BuildingDetails   `json:"building_details"`
		CurrentOwners    []UnitOwner       `json:"current_owners"`
		OwnershipHistory []OwnershipRecord `json:"ownership_history"`
	}

	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		owners, err := cfg.Db.GetUnitOwners(req.Context(), database.GetUnitOwnersParams{
			UnitID:        unit.ID,
			AssociationID: building.AssociationID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit owners", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve unit owners")
			return
		}
		ownerships, err := cfg.Db.GetUnitOwnerships(req.Context(), database.GetUnitOwnershipsParams{
			UnitID:        unit.ID,
			AssociationID: building.AssociationID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting unit ownerships", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve unit ownership history")
			return
		}

		response := UnitReportResponse{
			UnitDetails: UnitDetails{
				ID:              unit.ID,
				CadastralNumber: unit.CadastralNumber,
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
			},
			BuildingDetails: BuildingDetails{
				ID:              building.ID,
				Name:            building.Name,
				Address:         building.Address,
				CadastralNumber: building.CadastralNumber,
				TotalArea:       building.TotalArea,
			},
			CurrentOwners:    []UnitOwner{},
			OwnershipHistory: []OwnershipRecord{},
		}

		activeOwnerIDs := make(map[int64]bool)
		for _, o := range ownerships {
			if o.IsActive {
				activeOwnerIDs[o.OwnerID] = true
			}
		}
		// GetUnitOwners yields one row per ownership, so an owner with
		// several (e.g. a past and a current one) appears more than once.
		seen := make(map[int64]bool)
		for _, owner := range owners {
			if seen[owner.ID] {
				continue
			}
			seen[owner.ID] = true
			response.CurrentOwners = append(response.CurrentOwners, UnitOwner{
				ID:                   owner.ID,
				Name:                 owner.Name,
				IdentificationNumber: owner.IdentificationNumber,
				ContactPhone:         owner.ContactPhone,
				ContactEmail:         owner.ContactEmail,
				IsActive:             activeOwnerIDs[owner.ID],
			})
		}

		for _, o := range ownerships {
			record := OwnershipRecord{
				ID:                   o.ID,
				OwnerID:              o.OwnerID,
				OwnerName:            o.OwnerName,
				StartDate:            o.StartDate.Time,
				IsActive:             o.IsActive,
				IsVoting:             o.IsVoting,
				RegistrationDocument: o.RegistrationDocument,
				RegistrationDate:     o.RegistrationDate,
			}
			if o.EndDate.Valid {
				endDate := o.EndDate.Time
				record.EndDate = &endDate
			}
			response.OwnershipHistory = append(response.OwnershipHistory, record)
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}
