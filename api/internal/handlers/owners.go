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

type Ownership struct {
	ID                   int64     `json:"id"`
	UnitId               int64     `json:"unit_id"`
	OwnerId              int64     `json:"owner_id"`
	AssociationId        int64     `json:"association_id"`
	StartDate            time.Time `json:"start_date"`
	EndDate              time.Time `json:"end_date"`
	IsActive             bool      `json:"is_active"`
	RegistrationDocument string    `json:"registration_document"`
	RegistrationDate     time.Time `json:"registration_date"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func HandleGetAssoctiationOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		//buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		ownersFromDb, err := cfg.Db.GetAssociationOwners(req.Context(), int64(associationId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			log.Printf(errors)
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
