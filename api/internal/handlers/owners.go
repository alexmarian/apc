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
	ID                        int64     `json:"id"`
	UnitId                    int64     `json:"unit_id"`
	OwnerId                   int64     `json:"owner_id"`
	OwnerName                 string    `json:"owner_name"`
	OwnerNormalizedName       string    `json:"owner_normalized_name"`
	OwnerIdentificationNumber string    `json:"owner_identification_number"`
	AssociationId             int64     `json:"association_id"`
	StartDate                 time.Time `json:"start_date"`
	EndDate                   time.Time `json:"end_date"`
	IsActive                  bool      `json:"is_active"`
	RegistrationDocument      string    `json:"registration_document"`
	RegistrationDate          time.Time `json:"registration_date"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
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

		// Parse query parameters for specific owner ID (optional)
		var specificOwnerId int64 = 0

		if ownerIdStr := req.URL.Query().Get("owner_id"); ownerIdStr != "" {
			ownerId, err := strconv.ParseInt(ownerIdStr, 10, 64)
			if err != nil {
				RespondWithError(rw, http.StatusBadRequest, "Invalid owner_id parameter")
				return
			}
			specificOwnerId = ownerId
		}

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

		// Use the optimized query that filters at the database level
		reportData, err := cfg.Db.GetOwnerUnitsWithDetailsForReport(req.Context(), database.GetOwnerUnitsWithDetailsForReportParams{
			AssociationID: int64(associationId),
			ID:            specificOwnerId,
		})

		if err != nil {
			log.Printf("Error retrieving owner report data: %s", err)
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve owner report data")
			return
		}

		// Process the data
		ownerReports := []OwnerReportItem{}
		ownerMap := make(map[int64]*OwnerReportItem)
		processedOwnerIDs := make(map[int64]bool)

		// First pass: Create owner entries and collect units
		for _, row := range reportData {
			// Skip owners that have already been processed as co-owners
			if processedOwnerIDs[row.OwnerID] {
				continue
			}

			// If this owner isn't in our map yet, create a new entry
			if _, exists := ownerMap[row.OwnerID]; !exists {
				ownerMap[row.OwnerID] = &OwnerReportItem{
					Owner: Owner{
						ID:                   row.OwnerID,
						Name:                 row.OwnerName,
						NormalizedName:       row.OwnerNormalizedName,
						IdentificationNumber: row.OwnerIdentificationNumber,
						ContactPhone:         row.OwnerContactPhone,
						ContactEmail:         row.OwnerContactEmail,
						FirstDetectedAt:      row.OwnerFirstDetectedAt.Time,
						CreatedAt:            row.OwnerCreatedAt.Time,
						UpdatedAt:            row.OwnerUpdatedAt.Time,
					},
					Statistics: OwnerStats{},
					Units:      []OwnerUnit{},
					CoOwners:   []CoOwner{},
				}
			}

			// Track unique units for statistics
			uniqueUnitIDs := make(map[int64]bool)
			ownerEntry := ownerMap[row.OwnerID]

			// If this unit isn't already counted
			if !uniqueUnitIDs[row.UnitID] {
				uniqueUnitIDs[row.UnitID] = true

				// Update statistics
				ownerEntry.Statistics.TotalUnits++
				ownerEntry.Statistics.TotalArea += row.Area
				ownerEntry.Statistics.TotalCondoPart += row.Part

				// Include unit details if requested
				if includeUnits {
					ownerEntry.Units = append(ownerEntry.Units, OwnerUnit{
						UnitID:          row.UnitID,
						UnitNumber:      row.UnitNumber,
						BuildingName:    row.BuildingName,
						BuildingAddress: row.BuildingAddress,
						Area:            row.Area,
						Part:            row.Part,
						UnitType:        row.UnitType,
					})
				}
			}

			// Process co-owners if requested and available
			if includeCoOwners && row.CoOwnerID.Valid && row.CoOwnerID.Int64 != row.OwnerID {
				coOwnerID := row.CoOwnerID.Int64

				// Check if this co-owner already exists in the list
				coOwnerExists := false
				for i, coOwner := range ownerEntry.CoOwners {
					if coOwner.ID == coOwnerID {
						coOwnerExists = true

						// Add unit to existing co-owner if not already present
						unitExists := false
						for _, unitID := range ownerEntry.CoOwners[i].SharedUnitIDs {
							if unitID == row.UnitID {
								unitExists = true
								break
							}
						}

						if !unitExists {
							ownerEntry.CoOwners[i].SharedUnitIDs = append(
								ownerEntry.CoOwners[i].SharedUnitIDs,
								row.UnitID,
							)
						}
						break
					}
				}

				// Add new co-owner if not already in the list
				if !coOwnerExists {
					ownerEntry.CoOwners = append(ownerEntry.CoOwners, CoOwner{
						Owner: Owner{
							ID:                   coOwnerID,
							Name:                 row.CoOwnerName.String,
							NormalizedName:       row.CoOwnerNormalizedName.String,
							IdentificationNumber: row.CoOwnerIdentificationNumber.String,
							ContactPhone:         row.CoOwnerContactPhone.String,
							ContactEmail:         row.CoOwnerContactEmail.String,
						},
						SharedUnitIDs: []int64{row.UnitID},
					})
				}

				// Mark this co-owner as processed
				processedOwnerIDs[coOwnerID] = true
			}
		}

		// Convert map to slice
		for _, ownerEntry := range ownerMap {
			// Check if we should include these optional fields
			if !includeUnits {
				ownerEntry.Units = nil
			}
			if !includeCoOwners {
				ownerEntry.CoOwners = nil
			}

			ownerReports = append(ownerReports, *ownerEntry)
		}

		// Sort the ownerReports slice by total_condo_part in descending order
		sort.Slice(ownerReports, func(i, j int) bool {
			return ownerReports[i].Statistics.TotalCondoPart > ownerReports[j].Statistics.TotalCondoPart
		})

		RespondWithJSON(rw, http.StatusOK, ownerReports)
	}
}
