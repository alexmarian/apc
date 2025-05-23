package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

const OwnerIdPathValue = "ownerId"
const OwnershipIdPathValue = "ownershipId"

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

type OwnerUpdateRequest struct {
	Name                 string `json:"name"`
	NormalizedName       string `json:"normalized_name"`
	IdentificationNumber string `json:"identification_number"`
	ContactPhone         string `json:"contact_phone"`
	ContactEmail         string `json:"contact_email"`
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
	IsVoting                  bool      `json:"is_voting"`
	RegistrationDocument      string    `json:"registration_document"`
	RegistrationDate          time.Time `json:"registration_date"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

const UnitTypeVotersFilter = "unit_types"
const EntranceVotersFilter = "entrances"
const FloorVotersFilter = "floors"

func HandleGetAssociationOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		//buildingId, _ := strconv.Atoi(req.PathValue(BuildingIdPathValue))
		ownersFromDb, err := cfg.Db.GetAssociationOwners(req.Context(), int64(associationId))
		if err != nil {
			var errors = fmt.Sprintf("Error getting associations: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting associations")
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

func HandleUpdateAssociationOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		ownerId, _ := strconv.Atoi(req.PathValue(OwnerIdPathValue))
		dbOwner, err := cfg.Db.GetAssociationOwner(req.Context(), database.GetAssociationOwnerParams{
			ID:            int64(ownerId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			var errors = fmt.Sprintf("Error getting association owner: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting association owner")
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		var ownerUpdateRequest OwnerUpdateRequest
		if err := decoder.Decode(&ownerUpdateRequest); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		applyOwnerPartialUpdates(&ownerUpdateRequest, &dbOwner)
		err = cfg.Db.UpdateAssociationOwner(req.Context(), database.UpdateAssociationOwnerParams{
			Name:                 ownerUpdateRequest.Name,
			NormalizedName:       ownerUpdateRequest.NormalizedName,
			IdentificationNumber: ownerUpdateRequest.IdentificationNumber,
			ContactPhone:         ownerUpdateRequest.ContactPhone,
			ContactEmail:         ownerUpdateRequest.ContactEmail,
			ID:                   int64(ownerId),
			AssociationID:        int64(associationId),
		})
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, Owner{
			ID:                   dbOwner.ID,
			Name:                 dbOwner.Name,
			NormalizedName:       dbOwner.NormalizedName,
			IdentificationNumber: dbOwner.IdentificationNumber,
			ContactPhone:         dbOwner.ContactPhone,
			ContactEmail:         dbOwner.ContactEmail,
			FirstDetectedAt:      dbOwner.FirstDetectedAt.Time,
			CreatedAt:            dbOwner.CreatedAt.Time,
			UpdatedAt:            time.Now(),
		})
	}
}
func applyOwnerPartialUpdates(update *OwnerUpdateRequest, existing *database.Owner) {
	// Use pointers to distinguish between "field not provided" and "field set to zero value"
	if update.Name != "" {
		existing.Name = update.Name
		existing.NormalizedName = strings.ToLower(update.Name)
	} else {
		update.Name = existing.Name
		update.NormalizedName = existing.NormalizedName
	}
	if update.IdentificationNumber != "" {
		existing.IdentificationNumber = update.IdentificationNumber
	} else {
		update.IdentificationNumber = existing.IdentificationNumber
	}
	if update.ContactPhone != "" {
		existing.ContactPhone = update.ContactPhone
	} else {
		update.ContactPhone = existing.ContactPhone
	}
	if update.ContactEmail != "" {
		existing.ContactEmail = update.ContactEmail
	} else {
		update.ContactEmail = existing.ContactEmail
	}

}
func HandleGetAssociationOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		ownerId, _ := strconv.Atoi(req.PathValue(OwnerIdPathValue))
		dbOwner, err := cfg.Db.GetAssociationOwner(req.Context(), database.GetAssociationOwnerParams{
			ID:            int64(ownerId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			var errors = fmt.Sprintf("Error getting association owner: %s", err)
			logging.Logger.Log(zap.WarnLevel, "Error getting association owner")
			RespondWithError(rw, http.StatusInternalServerError, errors)
			return
		}
		RespondWithJSON(rw, http.StatusCreated, Owner{
			ID:                   dbOwner.ID,
			Name:                 dbOwner.Name,
			NormalizedName:       dbOwner.NormalizedName,
			IdentificationNumber: dbOwner.IdentificationNumber,
			ContactPhone:         dbOwner.ContactPhone,
			ContactEmail:         dbOwner.ContactEmail,
			FirstDetectedAt:      dbOwner.FirstDetectedAt.Time,
			CreatedAt:            dbOwner.CreatedAt.Time,
			UpdatedAt:            dbOwner.UpdatedAt.Time,
		})
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
			logging.Logger.Log(zap.WarnLevel, "Error creating owner", zap.String("error", err.Error()))
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

func HandleGetUnitOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))
		ownershipId, _ := strconv.Atoi(req.PathValue(OwnershipIdPathValue))

		ownership, err := cfg.Db.GetUnitOwnership(req.Context(), database.GetUnitOwnershipParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
			ID:            int64(ownershipId),
		})
		if err != nil && err != sql.ErrNoRows {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get current ownerships")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, Ownership{
			ID:                   ownership.ID,
			UnitId:               ownership.UnitID,
			OwnerId:              ownership.OwnerID,
			AssociationId:        ownership.AssociationID,
			StartDate:            ownership.StartDate.Time,
			EndDate:              ownership.EndDate.Time,
			IsActive:             ownership.IsActive,
			IsVoting:             ownership.IsVoting,
			RegistrationDocument: ownership.RegistrationDocument,
			CreatedAt:            ownership.CreatedAt.Time,
			UpdatedAt:            ownership.UpdatedAt.Time,
		})
	}
}

func HandleUnitVotingOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		unitId, _ := strconv.Atoi(req.PathValue(UnitIdPathValue))
		ownershipId, _ := strconv.Atoi(req.PathValue(OwnershipIdPathValue))
		err := cfg.Db.DisableActiveVoting(req.Context(), database.DisableActiveVotingParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
		})
		if err != nil && err != sql.ErrNoRows {
			logging.Logger.Log(zap.WarnLevel, "Error disabling active voting", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to disable active voting")
			return
		}
		err = cfg.Db.SetVoting(req.Context(), database.SetVotingParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),

			ID: int64(ownershipId),
		})
		if err != nil && err != sql.ErrNoRows {
			logging.Logger.Log(zap.WarnLevel, "Error setting active voting", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to set active voting")
			return
		}
		ownership, err := cfg.Db.GetUnitOwnership(req.Context(), database.GetUnitOwnershipParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
			ID:            int64(ownershipId),
		})
		if err != nil && err != sql.ErrNoRows {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get active voting")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, Ownership{
			ID:                   ownership.ID,
			UnitId:               ownership.UnitID,
			OwnerId:              ownership.OwnerID,
			AssociationId:        ownership.AssociationID,
			StartDate:            ownership.StartDate.Time,
			EndDate:              ownership.EndDate.Time,
			IsActive:             ownership.IsActive,
			IsVoting:             ownership.IsVoting,
			RegistrationDocument: ownership.RegistrationDocument,
			CreatedAt:            ownership.CreatedAt.Time,
			UpdatedAt:            ownership.UpdatedAt.Time,
		})
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
			ExclusiveOwnership   bool       `json:"is_exclusive"`
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
		currentOwnerships, err := cfg.Db.GetActiveUnitOwnerships(req.Context(), database.GetActiveUnitOwnershipsParams{
			UnitID:        int64(unitId),
			AssociationID: int64(associationId),
		})
		if err != nil && err != sql.ErrNoRows {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get current ownerships")
			return
		}

		if ownershipRequest.ExclusiveOwnership {
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
			IsVoting:             ownershipRequest.ExclusiveOwnership,
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
		includeUnits := req.URL.Query().Get("units") == "true"
		includeCoOwners := req.URL.Query().Get("co_owners") == "true"

		type OwnerUnit struct {
			UnitID              int64   `json:"unit_id"`
			UnitNumber          string  `json:"unit_number"`
			BuildingName        string  `json:"building_name"`
			UnitAddress         string  `json:"unit_address"`
			UnitCadastralNumber string  `json:"unit_cadastral_number"`
			Area                float64 `json:"area"`
			Part                float64 `json:"part"`
			UnitType            string  `json:"unit_type"`
		}

		type OwnerStats struct {
			TotalUnits     int     `json:"total_units"`
			TotalArea      float64 `json:"total_area"`
			TotalCondoPart float64 `json:"total_condo_part"`
		}

		type CoOwner struct {
			Owner
			SharedUnitIDs  []int64  `json:"shared_unit_ids"`
			SharedUnitNums []string `json:"shared_unit_nums"`
		}

		type OwnerReportItem struct {
			Owner      Owner       `json:"owner"`
			CoOwners   []CoOwner   `json:"co_owners,omitempty"`
			Units      []OwnerUnit `json:"units,omitempty"`
			Statistics OwnerStats  `json:"statistics"`
		}

		reportData, err := cfg.Db.GetOwnerUnitsWithDetailsForReport(req.Context(), database.GetOwnerUnitsWithDetailsForReportParams{
			AssociationID: int64(associationId),
			ID:            specificOwnerId,
			Column2:       specificOwnerId,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error retrieving owner report data", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve owner report data")
			return
		}

		ownerReports := []OwnerReportItem{}
		ownerMap := make(map[int64]*OwnerReportItem)

		for _, row := range reportData {
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

			ownerEntry := ownerMap[row.OwnerID]

			// Track unique units for statistics
			uniqueUnitIDs := make(map[int64]bool)
			if !uniqueUnitIDs[row.UnitID] {
				uniqueUnitIDs[row.UnitID] = true

				// Update statistics
				ownerEntry.Statistics.TotalUnits++
				ownerEntry.Statistics.TotalArea += row.Area
				ownerEntry.Statistics.TotalCondoPart += row.Part

				// Include unit details if requested
				if includeUnits {
					ownerEntry.Units = append(ownerEntry.Units, OwnerUnit{
						UnitID:              row.UnitID,
						UnitNumber:          row.UnitNumber,
						BuildingName:        row.BuildingName,
						UnitAddress:         row.UnitAddress,
						UnitCadastralNumber: row.UnitCadastralNumber,
						Area:                row.Area,
						Part:                row.Part,
						UnitType:            row.UnitType,
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
							ownerEntry.CoOwners[i].SharedUnitNums = append(
								ownerEntry.CoOwners[i].SharedUnitNums,
								row.UnitNumber,
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
						SharedUnitIDs:  []int64{row.UnitID},
						SharedUnitNums: []string{row.UnitNumber},
					})
				}
			}
		}

		// Convert map to slice
		for _, ownerEntry := range ownerMap {
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
func HandleGetVotersReport(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		unitTypes := strArrayToRelevantStrArray(strings.Split(req.URL.Query().Get(UnitTypeVotersFilter), ","))
		entrance, _ := strArrayToInt64Array(strings.Split(req.URL.Query().Get(EntranceVotersFilter), ","))
		floor, _ := strArrayToInt64Array(strings.Split(req.URL.Query().Get(FloorVotersFilter), ","))
		type OwnerUnit struct {
			UnitID          int64   `json:"unit_id"`
			UnitNumber      string  `json:"unit_number"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			Area            float64 `json:"area"`
			Part            float64 `json:"part"`
			UnitType        string  `json:"unit_type"`
		}

		type OwnerReportItem struct {
			OwnerId              int64       `json:"owner_id"`
			OwnerName            string      `json:"name"`
			IdentificationNumber string      `json:"identification_number"`
			ContactPhone         string      `json:"contact_phone"`
			ContactEmail         string      `json:"contact_email"`
			Units                []OwnerUnit `json:"units,omitempty"`
			TotalUnits           int         `json:"total_units"`
			TotalArea            float64     `json:"total_area"`
			VotingShare          float64     `json:"total_condo_part"`
		}

		voterData, err := cfg.Db.GetAssociationVoters(req.Context(),
			database.GetAssociationVotersParams{
				AssociationID: int64(associationId),
				Column2:       len(unitTypes) > 0,
				UnitTypes:     unitTypes,
				Column4:       len(entrance) > 0,
				UnitEntrances: entrance,
				Column6:       len(floor) > 0,
				UnitFloors:    floor,
			})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error retrieving owner report data", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to retrieve owner report data")
			return
		}

		ownerReports := []OwnerReportItem{}
		ownerMap := make(map[int64]*OwnerReportItem)

		for _, row := range voterData {
			// If this owner isn't in our map yet, create a new entry
			if _, exists := ownerMap[row.OwnerID]; !exists {
				ownerMap[row.OwnerID] = &OwnerReportItem{

					OwnerId:              row.OwnerID,
					OwnerName:            row.OwnerName,
					IdentificationNumber: row.OwnerIdentificationNumber,
					ContactPhone:         row.OwnerContactPhone,
					ContactEmail:         row.OwnerContactEmail,
					TotalArea:            0,
					VotingShare:          0,
					TotalUnits:           0,
					Units:                []OwnerUnit{},
				}
			}

			ownerEntry := ownerMap[row.OwnerID]

			ownerEntry.Units = append(ownerEntry.Units, OwnerUnit{
				UnitID:          row.UnitID,
				UnitNumber:      row.UnitNumber,
				BuildingName:    row.BuildingName,
				BuildingAddress: row.BuildingAddress,
				Area:            row.Area,
				Part:            row.Part,
				UnitType:        row.UnitType,
			})
			ownerEntry.TotalArea += row.Area
			ownerEntry.VotingShare += row.Part
			ownerEntry.TotalUnits++
		}
		for _, ownerEntry := range ownerMap {
			ownerReports = append(ownerReports, *ownerEntry)
		}
		RespondWithJSON(rw, http.StatusOK, ownerReports)
	}
}

func strArrayToInt64Array(strArray []string) ([]int64, error) {
	intArray := make([]int64, len(strArray))
	for i, str := range strArray {
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return nil, err
		}
		intArray[i] = val
	}
	return intArray, nil
}

func strArrayToRelevantStrArray(strArray []string) []string {
	relArray := []string{}
	for _, str := range strArray {
		if len(str) > 0 {
			relArray = append(relArray, str)
		}
	}
	return relArray
}
