package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

func HandleGetAssociationOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
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
func HandleCreateOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse request
		decoder := json.NewDecoder(req.Body)
		var newOwner struct {
			Name                 string `json:"name"`
			IdentificationNumber string `json:"identification_number"`
			ContactPhone         string `json:"contact_phone"`
			ContactEmail         string `json:"contact_email"`
		}

		if err := decoder.Decode(&newOwner); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate
		if newOwner.Name == "" {
			RespondWithError(rw, http.StatusBadRequest, "Owner name is required")
			return
		}

		// Normalize name for search
		normalizedName := strings.ToLower(newOwner.Name)

		// Create owner in database
		owner, err := cfg.Db.CreateOwner(req.Context(), database.CreateOwnerParams{
			Name:                 newOwner.Name,
			NormalizedName:       normalizedName,
			IdentificationNumber: newOwner.IdentificationNumber,
			ContactPhone:         newOwner.ContactPhone,
			ContactEmail:         newOwner.ContactEmail,
			AssociationID:        int64(associationId),
		})

		if err != nil {
			log.Printf("Error creating owner: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create owner")
			return
		}

		// Return the created owner
		ownerResponse := Owner{
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

		RespondWithJSON(rw, http.StatusCreated, ownerResponse)
	}
}
func HandleCreateUnitOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))

		// Parse request
		decoder := json.NewDecoder(req.Body)
		var ownershipRequest struct {
			OwnerID              int64      `json:"owner_id"`
			StartDate            time.Time  `json:"start_date"`
			EndDate              *time.Time `json:"end_date,omitempty"`
			RegistrationDocument string     `json:"registration_document"`
			RegistrationDate     time.Time  `json:"registration_date"`
		}

		if err := decoder.Decode(&ownershipRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Check if owner exists
		_, err := cfg.Db.GetOwnerById(req.Context(), ownershipRequest.OwnerID)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		// Get current active ownerships for this unit
		currentOwnerships, err := cfg.Db.GetActiveUnitOwnerships(req.Context(), int64(unitId))
		if err != nil && err != sql.ErrNoRows {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get current ownerships")
			return
		}

		// Mark current ownerships as inactive
		for _, ownership := range currentOwnerships {
			now := time.Now()
			err = cfg.Db.DeactivateOwnership(req.Context(), database.DeactivateOwnershipParams{
				ID:      ownership.ID,
				EndDate: sql.NullTime{Time: now, Valid: true},
			})

			if err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Failed to deactivate current ownership")
				return
			}
		}

		// Create new ownership
		var endDateParam sql.NullTime
		if ownershipRequest.EndDate != nil {
			endDateParam = sql.NullTime{Time: *ownershipRequest.EndDate, Valid: true}
		}

		newOwnership, err := cfg.Db.CreateOwnership(req.Context(), database.CreateOwnershipParams{
			UnitID:               int64(unitId),
			OwnerID:              ownershipRequest.OwnerID,
			AssociationID:        int64(associationId),
			StartDate:            sql.NullTime{Time: ownershipRequest.StartDate},
			EndDate:              endDateParam,
			IsActive:             true,
			RegistrationDocument: ownershipRequest.RegistrationDocument,
			RegistrationDate:     ownershipRequest.RegistrationDate,
		})

		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create new ownership")
			return
		}

		// Return the new ownership
		ownershipResponse := Ownership{
			ID:                   newOwnership.ID,
			UnitId:               newOwnership.UnitID,
			OwnerId:              newOwnership.OwnerID,
			AssociationId:        newOwnership.AssociationID,
			StartDate:            newOwnership.StartDate.Time,
			EndDate:              endDateParam.Time,
			IsActive:             newOwnership.IsActive,
			RegistrationDocument: newOwnership.RegistrationDocument,
			RegistrationDate:     newOwnership.RegistrationDate,
			CreatedAt:            newOwnership.CreatedAt.Time,
			UpdatedAt:            newOwnership.UpdatedAt.Time,
		}

		RespondWithJSON(rw, http.StatusCreated, ownershipResponse)
	}
}
func HandleGetOwnerReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		// Parse query parameters for data inclusion options
		includeUnits := false
		includeCoOwners := false

		if req.URL.Query().Get("units") == "true" {
			includeUnits = true
		}

		if req.URL.Query().Get("co_owners") == "true" {
			includeCoOwners = true
		}

		// Define response types
		type OwnerUnit struct {
			UnitID          int64   `json:"unit_id"`
			UnitNumber      string  `json:"unit_number"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			Area            float64 `json:"area"`
			Part            float64 `json:"part"`
			UnitType        string  `json:"unit_type"`
		}

		type OwnerStats struct {
			TotalUnits     int     `json:"total_units"`
			TotalArea      float64 `json:"total_area"`
			TotalCondoPart float64 `json:"total_condo_part"`
		}

		// Enhanced co-owner type with shared units
		type CoOwner struct {
			Owner
			SharedUnitIDs []int64 `json:"shared_unit_ids"` // The unit IDs they co-own
		}

		type OwnerReportItem struct {
			Owner      Owner       `json:"owner"`
			CoOwners   []CoOwner   `json:"co_owners,omitempty"`
			Units      []OwnerUnit `json:"units,omitempty"`
			Statistics OwnerStats  `json:"statistics"`
		}

		// Get all owners for the association
		ownersFromDb, err := cfg.Db.GetAssociationOwners(req.Context(), int64(associationId))
		if err != nil {
			log.Printf("Error getting owners: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve owners")
			return
		}

		ownerReports := []OwnerReportItem{}
		processedOwnerIDs := make(map[int64]bool)

		for _, ownerDb := range ownersFromDb {
			// Skip already processed owners
			if processedOwnerIDs[ownerDb.ID] {
				continue
			}
			processedOwnerIDs[ownerDb.ID] = true

			owner := Owner{
				ID:                   ownerDb.ID,
				Name:                 ownerDb.Name,
				NormalizedName:       ownerDb.NormalizedName,
				IdentificationNumber: ownerDb.IdentificationNumber,
				ContactPhone:         ownerDb.ContactPhone,
				ContactEmail:         ownerDb.ContactEmail,
				FirstDetectedAt:      ownerDb.FirstDetectedAt.Time,
				CreatedAt:            ownerDb.CreatedAt.Time,
				UpdatedAt:            ownerDb.UpdatedAt.Time,
			}

			// Get owner's units with details
			unitsWithDetails, err := cfg.Db.GetOwnerUnitsWithDetails(req.Context(), database.GetOwnerUnitsWithDetailsParams{
				OwnerID:       ownerDb.ID,
				AssociationID: int64(associationId),
			})

			if err != nil {
				log.Printf("Error getting units for owner %d: %s", ownerDb.ID, err)
				continue
			}

			// Process units
			ownerUnits := []OwnerUnit{}
			stats := OwnerStats{}

			uniqueUnitIDs := make(map[int64]bool)

			// Map to track co-owners and their shared units
			coOwnerMap := make(map[int64]*CoOwner)

			for _, unit := range unitsWithDetails {
				// Only process each unit once for statistics
				if !uniqueUnitIDs[unit.UnitID] {
					uniqueUnitIDs[unit.UnitID] = true
					stats.TotalUnits++
					stats.TotalArea += unit.Area
					stats.TotalCondoPart += unit.Part

					// Only collect unit details if requested
					if includeUnits {
						ownerUnits = append(ownerUnits, OwnerUnit{
							UnitID:          unit.UnitID,
							UnitNumber:      unit.UnitNumber,
							BuildingName:    unit.BuildingName,
							BuildingAddress: unit.BuildingAddress,
							Area:            unit.Area,
							Part:            unit.Part,
							UnitType:        unit.UnitType,
						})
					}
				}

				// Collect co-owners with their shared unit IDs
				if includeCoOwners && unit.CoOwnerID.Valid && unit.CoOwnerID.Int64 != ownerDb.ID {
					coOwnerID := unit.CoOwnerID.Int64

					// If this is the first time we're seeing this co-owner
					if _, exists := coOwnerMap[coOwnerID]; !exists {
						coOwnerMap[coOwnerID] = &CoOwner{
							Owner: Owner{
								ID:                   coOwnerID,
								Name:                 unit.CoOwnerName.String,
								NormalizedName:       unit.CoOwnerNormalizedName.String,
								IdentificationNumber: unit.CoOwnerIdentificationNumber.String,
								ContactPhone:         unit.CoOwnerContactPhone.String,
								ContactEmail:         unit.CoOwnerContactEmail.String,
							},
							SharedUnitIDs: []int64{},
						}
					}

					// Add this unit ID to the co-owner's shared units
					coOwnerMap[coOwnerID].SharedUnitIDs = append(coOwnerMap[coOwnerID].SharedUnitIDs, unit.UnitID)
					processedOwnerIDs[coOwnerID] = true
				}
			}

			// Prepare report item
			reportItem := OwnerReportItem{
				Owner:      owner,
				Statistics: stats,
			}

			// Only add units if requested
			if includeUnits {
				reportItem.Units = ownerUnits
			}

			// Only add co-owners if requested
			if includeCoOwners {
				// Convert map to slice
				coOwners := []CoOwner{}
				for _, coOwner := range coOwnerMap {
					coOwners = append(coOwners, *coOwner)
				}

				reportItem.CoOwners = coOwners
			}

			// Add to report
			ownerReports = append(ownerReports, reportItem)
		}

		// Sort the ownerReports slice by total_condo_part in descending order
		sort.Slice(ownerReports, func(i, j int) bool {
			return ownerReports[i].Statistics.TotalCondoPart > ownerReports[j].Statistics.TotalCondoPart
		})

		RespondWithJSON(rw, http.StatusOK, ownerReports)
	}
}
