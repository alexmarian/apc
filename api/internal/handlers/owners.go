package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
	"net/http"
	"regexp"
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
	PhotoURL             string    `json:"photo_url,omitempty"`
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
	OwnerPhotoURL             string    `json:"owner_photo_url,omitempty"`
}

const UnitTypeVotersFilter = "unit_types"
const EntranceVotersFilter = "entrances"
const FloorVotersFilter = "floors"

func toOwner(o database.Owner) Owner {
	return Owner{
		ID:                   o.ID,
		Name:                 o.Name,
		NormalizedName:       o.NormalizedName,
		IdentificationNumber: o.IdentificationNumber,
		ContactPhone:         o.ContactPhone,
		ContactEmail:         o.ContactEmail,
		FirstDetectedAt:      o.FirstDetectedAt.Time,
		CreatedAt:            o.CreatedAt.Time,
		UpdatedAt:            o.UpdatedAt.Time,
	}
}

func HandleGetAssociationOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		ownersFromDb, err := cfg.Db.GetAssociationOwners(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting owners", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Error getting owners")
			return
		}
		owners := make([]Owner, len(ownersFromDb))
		photos := OwnerPhotoURLs(req.Context(), cfg, int64(associationId))
		for i, owner := range ownersFromDb {
			owners[i] = toOwner(owner)
			owners[i].PhotoURL = photos[owner.ID]
		}
		RespondWithJSON(rw, http.StatusOK, owners)
	}
}

// associationOwner resolves {ownerId} within the association in the path.
// On failure it has already written the response.
func associationOwner(cfg *ApiConfig, rw http.ResponseWriter, req *http.Request) (database.Owner, bool) {
	associationId, okA := pathInt64(req, AssociationIdPathValue)
	ownerId, okO := pathInt64(req, OwnerIdPathValue)
	if !okA || !okO {
		RespondWithError(rw, http.StatusBadRequest, "Invalid id")
		return database.Owner{}, false
	}
	owner, err := cfg.Db.GetAssociationOwner(req.Context(), database.GetAssociationOwnerParams{ID: ownerId, AssociationID: associationId})
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(rw, http.StatusNotFound, "Owner not found")
		return owner, false
	}
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting owner", zap.String("error", err.Error()))
		RespondWithError(rw, http.StatusInternalServerError, "Error getting owner")
		return owner, false
	}
	return owner, true
}

// ownerInput is a validated create/update body.
type ownerInput struct {
	Name                 string `json:"name"`
	IdentificationNumber string `json:"identification_number"`
	ContactPhone         string `json:"contact_phone"`
	ContactEmail         string `json:"contact_email"`
}

// parseOwnerInput decodes and validates an owner body: everything trimmed,
// name required, each field at most 255 characters, an email (if any) must
// look like one. On failure it has already written the response.
func parseOwnerInput(rw http.ResponseWriter, req *http.Request) (ownerInput, bool) {
	var in ownerInput
	if err := json.NewDecoder(req.Body).Decode(&in); err != nil {
		RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
		return in, false
	}
	for _, f := range []*string{&in.Name, &in.IdentificationNumber, &in.ContactPhone, &in.ContactEmail} {
		*f = strings.TrimSpace(*f)
		if len([]rune(*f)) > 255 {
			RespondWithError(rw, http.StatusBadRequest, "Fields must be at most 255 characters")
			return in, false
		}
	}
	switch {
	case in.Name == "":
		RespondWithError(rw, http.StatusBadRequest, "Owner name is required")
		return in, false
	case in.ContactEmail != "" && !emailPattern.MatchString(in.ContactEmail):
		RespondWithError(rw, http.StatusBadRequest, "Invalid email address")
		return in, false
	}
	return in, true
}

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// HandleUpdateAssociationOwner replaces an owner's details (all fields —
// an empty value clears it). 404 for an owner outside the association.
func HandleUpdateAssociationOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		in, ok := parseOwnerInput(rw, req)
		if !ok {
			return
		}
		err := cfg.Db.UpdateAssociationOwner(req.Context(), database.UpdateAssociationOwnerParams{
			Name:                 in.Name,
			NormalizedName:       strings.ToLower(in.Name),
			IdentificationNumber: in.IdentificationNumber,
			ContactPhone:         in.ContactPhone,
			ContactEmail:         in.ContactEmail,
			ID:                   owner.ID,
			AssociationID:        owner.AssociationID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating owner", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update owner")
			return
		}
		owner.Name, owner.NormalizedName = in.Name, strings.ToLower(in.Name)
		owner.IdentificationNumber, owner.ContactPhone, owner.ContactEmail = in.IdentificationNumber, in.ContactPhone, in.ContactEmail
		result := toOwner(owner)
		result.UpdatedAt = time.Now()
		RespondWithJSON(rw, http.StatusOK, result)
	}
}

func HandleGetAssociationOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		owner, ok := associationOwner(cfg, rw, req)
		if !ok {
			return
		}
		o := toOwner(owner)
		o.PhotoURL = OwnerPhotoURLs(req.Context(), cfg, owner.AssociationID)[owner.ID]
		RespondWithJSON(rw, http.StatusOK, o)
	}
}
func HandleCreateOwner(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		in, ok := parseOwnerInput(rw, req)
		if !ok {
			return
		}
		owner, err := cfg.Db.CreateOwner(req.Context(), database.CreateOwnerParams{
			Name:                 in.Name,
			NormalizedName:       strings.ToLower(in.Name),
			IdentificationNumber: in.IdentificationNumber,
			ContactPhone:         in.ContactPhone,
			ContactEmail:         in.ContactEmail,
			AssociationID:        int64(associationId),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating owner", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create owner")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, toOwner(owner))
	}
}

func toOwnership(o database.GetUnitOwnershipRow) Ownership {
	return Ownership{
		ID:                        o.ID,
		UnitId:                    o.UnitID,
		OwnerId:                   o.OwnerID,
		OwnerName:                 o.OwnerName,
		OwnerNormalizedName:       o.OwnerNormalizedName,
		OwnerIdentificationNumber: o.IdentificationNumber,
		AssociationId:             o.AssociationID,
		StartDate:                 o.StartDate.Time,
		EndDate:                   o.EndDate.Time,
		IsActive:                  o.IsActive,
		IsVoting:                  o.IsVoting,
		RegistrationDocument:      o.RegistrationDocument,
		RegistrationDate:          o.RegistrationDate,
		CreatedAt:                 o.CreatedAt.Time,
		UpdatedAt:                 o.UpdatedAt.Time,
	}
}

// unitOwnership resolves {ownershipId} within the unit in the path. On
// failure it has already written the response.
func unitOwnership(q *database.Queries, rw http.ResponseWriter, req *http.Request, building database.Building, unit database.Unit) (database.GetUnitOwnershipRow, bool) {
	ownershipId, ok := pathInt64(req, OwnershipIdPathValue)
	if !ok {
		RespondWithError(rw, http.StatusBadRequest, "Invalid ownership id")
		return database.GetUnitOwnershipRow{}, false
	}
	ownership, err := q.GetUnitOwnership(req.Context(), database.GetUnitOwnershipParams{
		UnitID:        unit.ID,
		AssociationID: building.AssociationID,
		ID:            ownershipId,
	})
	if errors.Is(err, sql.ErrNoRows) {
		RespondWithError(rw, http.StatusNotFound, "Ownership not found")
		return ownership, false
	}
	if err != nil {
		RespondWithError(rw, http.StatusInternalServerError, "Failed to get ownership")
		return ownership, false
	}
	return ownership, true
}

func HandleGetUnitOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		ownership, ok := unitOwnership(cfg.Db, rw, req, building, unit)
		if !ok {
			return
		}
		RespondWithJSON(rw, http.StatusOK, toOwnership(ownership))
	}
}

// HandleUnitVotingOwnership makes an active ownership the unit's voting
// one, taking voting rights from the previous holder — in one transaction,
// so a failure never leaves the unit without its voting owner.
func HandleUnitVotingOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		ctx := req.Context()
		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to set voting owner")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)

		ownership, ok := unitOwnership(q, rw, req, building, unit)
		if !ok {
			return
		}
		if !ownership.IsActive {
			RespondWithError(rw, http.StatusBadRequest, "Only an active ownership can have voting rights")
			return
		}
		err = q.DisableActiveVoting(ctx, database.DisableActiveVotingParams{UnitID: unit.ID, AssociationID: building.AssociationID})
		if err == nil {
			err = q.SetVoting(ctx, database.SetVotingParams{ID: ownership.ID, UnitID: unit.ID, AssociationID: building.AssociationID})
		}
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error setting voting owner", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to set voting owner")
			return
		}
		ownership.IsVoting = true
		RespondWithJSON(rw, http.StatusOK, toOwnership(ownership))
	}
}

// isUniqueViolation reports a SQLite UNIQUE constraint failure.
func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}

// HandleCreateUnitOwnership records an owner (of this association) owning
// the unit. With is_exclusive, every other active ownership of the unit
// ends now and the new one becomes the voting ownership — all in one
// transaction.
func HandleCreateUnitOwnership(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		building, unit, ok := associationUnit(cfg, rw, req)
		if !ok {
			return
		}
		var request struct {
			OwnerID              int64      `json:"owner_id"`
			StartDate            time.Time  `json:"start_date"`
			EndDate              *time.Time `json:"end_date,omitempty"`
			RegistrationDocument string     `json:"registration_document"`
			RegistrationDate     time.Time  `json:"registration_date"`
			ExclusiveOwnership   bool       `json:"is_exclusive"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}
		request.RegistrationDocument = strings.TrimSpace(request.RegistrationDocument)
		switch {
		case request.RegistrationDocument == "":
			RespondWithError(rw, http.StatusBadRequest, "Registration document is required")
			return
		case request.StartDate.IsZero():
			RespondWithError(rw, http.StatusBadRequest, "Start date is required")
			return
		case request.RegistrationDate.IsZero():
			RespondWithError(rw, http.StatusBadRequest, "Registration date is required")
			return
		case request.EndDate != nil && request.EndDate.Before(request.StartDate):
			RespondWithError(rw, http.StatusBadRequest, "End date can't be before the start date")
			return
		}

		ctx := req.Context()
		if _, err := cfg.Db.GetAssociationOwner(ctx, database.GetAssociationOwnerParams{
			ID:            request.OwnerID,
			AssociationID: building.AssociationID,
		}); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		tx, err := cfg.DB.BeginTx(ctx, nil)
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create ownership")
			return
		}
		defer tx.Rollback()
		q := cfg.Db.WithTx(tx)

		if request.ExclusiveOwnership {
			current, err := q.GetActiveUnitOwnerships(ctx, database.GetActiveUnitOwnershipsParams{
				UnitID:        unit.ID,
				AssociationID: building.AssociationID,
			})
			if err != nil {
				RespondWithError(rw, http.StatusInternalServerError, "Failed to get current ownerships")
				return
			}
			now := sql.NullTime{Time: time.Now().UTC(), Valid: true}
			for _, o := range current {
				if err := q.DeactivateOwnership(ctx, database.DeactivateOwnershipParams{ID: o.ID, EndDate: now}); err != nil {
					if isUniqueViolation(err) {
						RespondWithError(rw, http.StatusConflict, ownershipHistoryConflict)
						return
					}
					RespondWithError(rw, http.StatusInternalServerError, "Failed to deactivate current ownership")
					return
				}
			}
		}

		var endDate sql.NullTime
		if request.EndDate != nil {
			endDate = sql.NullTime{Time: *request.EndDate, Valid: true}
		}
		created, err := q.CreateOwnership(ctx, database.CreateOwnershipParams{
			UnitID:               unit.ID,
			OwnerID:              request.OwnerID,
			AssociationID:        building.AssociationID,
			StartDate:            sql.NullTime{Time: request.StartDate, Valid: true},
			EndDate:              endDate,
			IsActive:             true,
			IsVoting:             request.ExclusiveOwnership,
			RegistrationDocument: request.RegistrationDocument,
			RegistrationDate:     request.RegistrationDate,
		})
		if isUniqueViolation(err) {
			RespondWithError(rw, http.StatusConflict, "This owner already has an active ownership of this unit")
			return
		}
		if err == nil {
			err = tx.Commit()
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating ownership", zap.String("error", err.Error()))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create new ownership")
			return
		}

		row, err := cfg.Db.GetUnitOwnership(ctx, database.GetUnitOwnershipParams{
			UnitID:        unit.ID,
			AssociationID: building.AssociationID,
			ID:            created.ID,
		})
		if err != nil {
			RespondWithError(rw, http.StatusInternalServerError, "Failed to load new ownership")
			return
		}
		RespondWithJSON(rw, http.StatusCreated, toOwnership(row))
	}
}

// ownershipHistoryConflict explains a schema limit: ownerships are unique
// per (unit, owner, is_active), so one owner can have at most one past
// ownership of a given unit.
const ownershipHistoryConflict = "This owner already has a past ownership of this unit; it can't be ended again"

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
			BuildingID          int64   `json:"building_id"`
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
			Owner       Owner       `json:"owner"`
			CoOwners    []CoOwner   `json:"co_owners,omitempty"`
			Units       []OwnerUnit `json:"units,omitempty"`
			Statistics  OwnerStats  `json:"statistics"`
			seenUnitIDs map[int64]bool
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
					Statistics:  OwnerStats{},
					Units:       []OwnerUnit{},
					CoOwners:    []CoOwner{},
					seenUnitIDs: make(map[int64]bool),
				}
			}

			ownerEntry := ownerMap[row.OwnerID]

			// Track unique units for statistics
			if !ownerEntry.seenUnitIDs[row.UnitID] {
				ownerEntry.seenUnitIDs[row.UnitID] = true

				// Update statistics
				ownerEntry.Statistics.TotalUnits++
				ownerEntry.Statistics.TotalArea += row.Area
				ownerEntry.Statistics.TotalCondoPart += row.Part

				// Include unit details if requested
				if includeUnits {
					ownerEntry.Units = append(ownerEntry.Units, OwnerUnit{
						UnitID:              row.UnitID,
						UnitNumber:          row.UnitNumber,
						BuildingID:          row.BuildingID,
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
		photos := OwnerPhotoURLs(req.Context(), cfg, int64(associationId))
		for _, ownerEntry := range ownerMap {
			ownerEntry.Owner.PhotoURL = photos[ownerEntry.Owner.ID]
			for i := range ownerEntry.CoOwners {
				ownerEntry.CoOwners[i].PhotoURL = photos[ownerEntry.CoOwners[i].ID]
			}
			if !includeUnits {
				ownerEntry.Units = nil
			}
			if !includeCoOwners {
				ownerEntry.CoOwners = nil
			}
			ownerReports = append(ownerReports, *ownerEntry)
		}

		// Sort the ownerReports slice by total_condo_part in descending order
		// By share, largest first; ties by name so the order is stable.
		sort.Slice(ownerReports, func(i, j int) bool {
			a, b := ownerReports[i], ownerReports[j]
			if a.Statistics.TotalCondoPart != b.Statistics.TotalCondoPart {
				return a.Statistics.TotalCondoPart > b.Statistics.TotalCondoPart
			}
			return a.Owner.NormalizedName < b.Owner.NormalizedName
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
