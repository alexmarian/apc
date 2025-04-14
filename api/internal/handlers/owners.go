package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const OwnerIdPathValue = "ownerId"

type Owner struct {
	ID                   int64     `json:"id"`
	Name                 string    `json:"name"`
	NormalizedName       string    `json:"normalized_name"`
	IdentificationNumber string    `json:"identification_number"`
	ContactPhone         string    `json:"contact_phone"`
	ContactEmail         string    `json:"contact_email"`
	FirstDetectedAt      time.Time `json:"first_detected_at"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func HandleGetAssoctiationOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
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
