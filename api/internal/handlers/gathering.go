package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

const GatheringIdPathValue = "gatheringId"
const VotingMatterIdPathValue = "matterId"
const ParticipantIdPathValue = "participantId"

// Gathering types
type Gathering struct {
	ID                          int64     `json:"id"`
	AssociationID               int64     `json:"association_id"`
	Title                       string    `json:"title"`
	Description                 string    `json:"description"`
	Intent                      string    `json:"intent"`
	GatheringDate               time.Time `json:"gathering_date"`
	GatheringType               string    `json:"gathering_type"`
	Status                      string    `json:"status"`
	QualificationUnitTypes      []string  `json:"qualification_unit_types"`
	QualificationFloors         []int64   `json:"qualification_floors"`
	QualificationEntrances      []int64   `json:"qualification_entrances"`
	QualificationCustomRule     string    `json:"qualification_custom_rule"`
	QualifiedUnitsCount         int       `json:"qualified_units_count"`
	QualifiedUnitsTotalPart     float64   `json:"qualified_units_total_part"`
	QualifiedUnitsTotalArea     float64   `json:"qualified_units_total_area"`
	ParticipatingUnitsCount     int       `json:"participating_units_count"`
	ParticipatingUnitsTotalPart float64   `json:"participating_units_total_part"`
	ParticipatingUnitsTotalArea float64   `json:"participating_units_total_area"`
	CreatedAt                   time.Time `json:"created_at"`
	UpdatedAt                   time.Time `json:"updated_at"`
}

type CreateGatheringRequest struct {
	Title                   string    `json:"title"`
	Description             string    `json:"description"`
	Intent                  string    `json:"intent"`
	Location                string    `json:"location"`
	GatheringDate           time.Time `json:"gathering_date"`
	GatheringType           string    `json:"gathering_type"`
	QualificationUnitTypes  []string  `json:"qualification_unit_types"`
	QualificationFloors     []int64   `json:"qualification_floors"`
	QualificationEntrances  []int64   `json:"qualification_entrances"`
	QualificationCustomRule string    `json:"qualification_custom_rule"`
}

type VotingMatter struct {
	ID           int64        `json:"id"`
	GatheringID  int64        `json:"gathering_id"`
	OrderIndex   int          `json:"order_index"`
	Title        string       `json:"title"`
	Description  string       `json:"description"`
	MatterType   string       `json:"matter_type"`
	VotingConfig VotingConfig `json:"voting_config"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}

type VotingConfig struct {
	Type                    string         `json:"type"` // yes_no, multiple_choice, ranking
	Options                 []VotingOption `json:"options,omitempty"`
	RequiredMajority        string         `json:"required_majority"` // simple, supermajority, custom
	RequiredMajorityValue   float64        `json:"required_majority_value,omitempty"`
	Quorum                  float64        `json:"quorum"`
	AllowAbstention         bool           `json:"allow_abstention"`
	IsAnonymous             bool           `json:"is_anonymous"`
	ShowResultsDuringVoting bool           `json:"show_results_during_voting"`
}

type VotingOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type GatheringParticipant struct {
	ID                        int64      `json:"id"`
	GatheringID               int64      `json:"gathering_id"`
	ParticipantType           string     `json:"participant_type"`
	ParticipantName           string     `json:"participant_name"`
	ParticipantIdentification string     `json:"participant_identification"`
	OwnerID                   *int64     `json:"owner_id"`
	DelegatingOwnerID         *int64     `json:"delegating_owner_id"`
	DelegationDocumentRef     string     `json:"delegation_document_ref"`
	UnitsInfo                 []int64    `json:"units_info"`
	UnitsPart                 float64    `json:"units_part"`
	UnitsArea                 float64    `json:"units_area"`
	CheckInTime               *time.Time `json:"check_in_time"`
	HasVoted                  bool       `json:"has_voted"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

type Ballot struct {
	ID                 int64                 `json:"id"`
	GatheringID        int64                 `json:"gathering_id"`
	ParticipantID      int64                 `json:"participant_id"`
	BallotContent      map[string]BallotVote `json:"ballot_content"`
	BallotHash         string                `json:"ballot_hash"`
	SubmittedAt        time.Time             `json:"submitted_at"`
	SubmittedIP        string                `json:"submitted_ip"`
	SubmittedUserAgent string                `json:"submitted_user_agent"`
	IsValid            bool                  `json:"is_valid"`
}

type BallotVote struct {
	MatterID  int64  `json:"matter_id"`
	OptionID  string `json:"option_id,omitempty"`
	VoteValue string `json:"vote_value"` // yes, no, abstain, or option ID
}

type VoteResults struct {
	GatheringID int64              `json:"gathering_id"`
	Results     []VoteMatterResult `json:"results"`
	Summary     GatheringSummary   `json:"summary"`
}

type VoteMatterResult struct {
	MatterID       int64                  `json:"matter_id"`
	MatterTitle    string                 `json:"matter_title"`
	MatterType     string                 `json:"matter_type"`
	VotingConfig   VotingConfig           `json:"voting_config"`
	Tally          map[string]TallyResult `json:"tally"`
	TotalVoted     float64                `json:"total_voted"`
	TotalAbstained float64                `json:"total_abstained"`
	Passed         bool                   `json:"passed"`
}

type TallyResult struct {
	Count      int     `json:"count"`
	Weight     float64 `json:"weight"`
	Area       float64 `json:"area"`
	Percentage float64 `json:"percentage"`
}

type GatheringSummary struct {
	QualifiedUnits       int     `json:"qualified_units"`
	QualifiedWeight      float64 `json:"qualified_weight"`
	QualifiedArea        float64 `json:"qualified_area"`
	ParticipatingUnits   int     `json:"participating_units"`
	ParticipatingWeight  float64 `json:"participating_weight"`
	ParticipatingArea    float64 `json:"participating_area"`
	VotedUnits           int     `json:"voted_units"`
	VotedWeight          float64 `json:"voted_weight"`
	VotedArea            float64 `json:"voted_area"`
	ParticipationRate    float64 `json:"participation_rate"`
	VotingCompletionRate float64 `json:"voting_completion_rate"`
}

// Handlers

func HandleGetGatherings(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		gatherings, err := cfg.Db.GetGatherings(req.Context(), int64(associationId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gatherings", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get gatherings")
			return
		}

		response := make([]Gathering, len(gatherings))
		for i, g := range gatherings {
			response[i] = dbGatheringToResponse(g)
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

func HandleGetGathering(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			if err == sql.ErrNoRows {
				RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			} else {
				logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
				RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering")
			}
			return
		}

		RespondWithJSON(rw, http.StatusOK, dbGatheringToResponse(gathering))
	}
}

func HandleCreateGathering(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))

		var createReq CreateGatheringRequest
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate required fields
		if createReq.Title == "" || createReq.GatheringType == "" {
			RespondWithError(rw, http.StatusBadRequest, "Title and gathering type are required")
			return
		}

		if createReq.Location == "" {
			RespondWithError(rw, http.StatusBadRequest, "Location is required")
			return
		}

		// Convert arrays to JSON strings for storage
		unitTypesJSON, _ := json.Marshal(createReq.QualificationUnitTypes)
		floorsJSON, _ := json.Marshal(createReq.QualificationFloors)
		entrancesJSON, _ := json.Marshal(createReq.QualificationEntrances)

		gathering, err := cfg.Db.CreateGathering(req.Context(), database.CreateGatheringParams{
			AssociationID:           int64(associationId),
			Title:                   createReq.Title,
			Description:             createReq.Description,
			Intent:                  createReq.Intent,
			Location:                createReq.Location,
			GatheringDate:           createReq.GatheringDate,
			GatheringType:           createReq.GatheringType,
			Status:                  "draft",
			QualificationUnitTypes:  sql.NullString{String: string(unitTypesJSON), Valid: len(unitTypesJSON) > 2},
			QualificationFloors:     sql.NullString{String: string(floorsJSON), Valid: len(floorsJSON) > 2},
			QualificationEntrances:  sql.NullString{String: string(entrancesJSON), Valid: len(entrancesJSON) > 2},
			QualificationCustomRule: sql.NullString{String: createReq.QualificationCustomRule, Valid: createReq.QualificationCustomRule != ""},
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating gathering", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering")
			return
		}

		// Calculate qualified units
		qualifiedCount, qualifiedPart, qualifiedArea := updateGatheringStats(cfg, gathering.ID, int64(associationId))

		// Log audit
		cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: gathering.ID,
			EntityType:  "gathering",
			EntityID:    gathering.ID,
			Action:      "created",
			PerformedBy: sql.NullString{String: req.Context().Value("userID").(string), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: "{}", Valid: true},
		})
		err = syncUnitsSlots(req, cfg, int64(associationId), gathering.ID, createReq.QualificationUnitTypes, createReq.QualificationFloors, createReq.QualificationEntrances)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error syncing units slots", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to sync units slots")
			return
		}
		response := dbGatheringToResponse(gathering)
		response.QualifiedUnitsCount = qualifiedCount
		response.QualifiedUnitsTotalPart = qualifiedPart
		response.QualifiedUnitsTotalArea = qualifiedArea
		RespondWithJSON(rw, http.StatusCreated, response)
	}
}

func syncUnitsSlots(req *http.Request, cfg *ApiConfig, associationId int64, gatheringId int64, unitTypes []string, floors []int64, entrances []int64) error {
	units, err := cfg.Db.GetQualifiedUnits(req.Context(), database.GetQualifiedUnitsParams{
		AssociationID: associationId,
		Column2:       len(unitTypes) > 0,
		UnitTypes:     unitTypes,
		Column4:       len(floors) > 0,
		UnitFloors:    floors,
		Column6:       len(entrances) > 0,
		UnitEntrances: entrances,
	})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting qualified units", zap.Error(err))
		return err
	}
	for _, unit := range units {
		_, err := cfg.Db.CreateUnitSlot(req.Context(), database.CreateUnitSlotParams{
			GatheringID: gatheringId,
			UnitID:      unit.ID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating unit slot", zap.Error(err))
			return err
		}
	}
	return nil
}

func HandleUpdateGatheringStatus(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		var statusReq struct {
			Status string `json:"status"`
		}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&statusReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate status transition
		validStatuses := map[string]bool{
			"draft": true, "published": true, "active": true, "closed": true, "tallied": true,
		}
		if !validStatuses[statusReq.Status] {
			RespondWithError(rw, http.StatusBadRequest, "Invalid status")
			return
		}

		gathering, err := cfg.Db.UpdateGatheringStatus(req.Context(), database.UpdateGatheringStatusParams{
			Status:        statusReq.Status,
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error updating gathering status", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to update gathering status")
			return
		}

		// If closing the gathering, calculate final results
		if statusReq.Status == "closed" {
			go calculateFinalResults(cfg, int64(gatheringId))
		}

		RespondWithJSON(rw, http.StatusOK, dbGatheringToResponse(gathering))
	}
}

func HandleGetVotingMatters(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		matters, err := cfg.Db.GetVotingMatters(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting voting matters", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting matters")
			return
		}

		response := make([]VotingMatter, len(matters))
		for i, m := range matters {
			response[i] = dbVotingMatterToResponse(m)
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

func HandleCreateVotingMatter(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		draft := validateGatheringStateWithFetch(req, cfg, gatheringId, associationId, "draft")
		if !draft {
			RespondWithError(rw, http.StatusBadRequest, "Cannot create voting matter in non-draft gathering")
			return
		}
		var createReq VotingMatter
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Convert voting config to JSON
		configJSON, err := json.Marshal(createReq.VotingConfig)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid voting configuration")
			return
		}

		matter, err := cfg.Db.CreateVotingMatter(req.Context(), database.CreateVotingMatterParams{
			GatheringID:  int64(gatheringId),
			OrderIndex:   int64(createReq.OrderIndex),
			Title:        createReq.Title,
			Description:  sql.NullString{String: createReq.Description, Valid: createReq.Description != ""},
			MatterType:   createReq.MatterType,
			VotingConfig: string(configJSON),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating voting matter", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create voting matter")
			return
		}

		RespondWithJSON(rw, http.StatusCreated, dbVotingMatterToResponse(matter))
	}
}

func HandleUpdateVotingMatter(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		matterId, _ := strconv.Atoi(req.PathValue(VotingMatterIdPathValue))
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		draft := validateGatheringStateWithFetch(req, cfg, gatheringId, associationId, "draft")
		if !draft {
			RespondWithError(rw, http.StatusBadRequest, "Cannot update voting matter in non-draft gathering")
			return
		}

		var createReq VotingMatter
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&createReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Convert voting config to JSON
		configJSON, err := json.Marshal(createReq.VotingConfig)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid voting configuration")
			return
		}

		matter, err := cfg.Db.UpdateVotingMatter(req.Context(), database.UpdateVotingMatterParams{
			GatheringID:  int64(gatheringId),
			ID:           int64(matterId),
			Title:        createReq.Title,
			Description:  sql.NullString{String: createReq.Description, Valid: createReq.Description != ""},
			MatterType:   createReq.MatterType,
			VotingConfig: string(configJSON),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating voting matter", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create voting matter")
			return
		}

		RespondWithJSON(rw, http.StatusCreated, dbVotingMatterToResponse(matter))
	}
}

func HandleDeleteVotingMatter(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		matterId, _ := strconv.Atoi(req.PathValue(VotingMatterIdPathValue))
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		draft := validateGatheringStateWithFetch(req, cfg, gatheringId, associationId, "draft")
		if !draft {
			RespondWithError(rw, http.StatusBadRequest, "Cannot delete voting matter in non-draft gathering")
			return
		}
		err := cfg.Db.DeleteVotingMatter(req.Context(), database.DeleteVotingMatterParams{
			GatheringID: int64(gatheringId),
			ID:          int64(matterId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error deleting voting matter", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to delete voting matter")
			return
		}

		RespondWithJSON(rw, http.StatusOK, map[string]string{"result": "success"})
	}
}

func validateGatheringState(gathering database.Gathering, targetStatus string) bool {
	return gathering.Status == targetStatus
}

func validateGatheringStateWithFetch(req *http.Request, cfg *ApiConfig, gatheringId int, associationId int, targetStatus string) bool {
	gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
		ID:            int64(gatheringId),
		AssociationID: int64(associationId),
	})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
		return false
	}
	return validateGatheringState(gathering, targetStatus)
}

func HandleAddParticipant(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		active := validateGatheringStateWithFetch(req, cfg, gatheringId, associationId, "active")
		if !active {
			RespondWithError(rw, http.StatusBadRequest, "Cannot add a participant in a  non-active gathering")
			return
		}

		var addReq struct {
			ParticipantType       string  `json:"participant_type"`
			OwnerID               int64   `json:"owner_id"`
			UnitIDs               []int64 `json:"unit_ids"`
			DelegatingOwnerID     *int64  `json:"delegating_owner_id,omitempty"`
			DelegationDocumentRef string  `json:"delegation_document_ref,omitempty"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&addReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate that unit IDs are provided and valid
		if len(addReq.UnitIDs) == 0 {
			RespondWithError(rw, http.StatusBadRequest, "Unit IDs are required for participation")
			return
		}

		// Filter out null or zero unit IDs and validate
		validUnitIDs := make([]int64, 0)
		for _, unitID := range addReq.UnitIDs {
			if unitID > 0 {
				validUnitIDs = append(validUnitIDs, unitID)
			}
		}

		if len(validUnitIDs) == 0 {
			RespondWithError(rw, http.StatusBadRequest, "At least one valid unit ID is required for participation")
			return
		}

		// Use the filtered unit IDs
		addReq.UnitIDs = validUnitIDs
		var effectiveOwnerID int64 = 0
		var participantID string
		var delegatingOwnerID int64 = 0
		if addReq.ParticipantType == "owner" && addReq.OwnerID != 0 {
			effectiveOwnerID = addReq.OwnerID
		} else if addReq.ParticipantType == "delegate" && addReq.DelegatingOwnerID != nil {
			effectiveOwnerID = *addReq.DelegatingOwnerID
			delegatingOwnerID = *addReq.DelegatingOwnerID
			participantID = addReq.DelegationDocumentRef
		} else {
			RespondWithError(rw, http.StatusBadRequest, "Invalid participant type or owner ID")
			return
		}
		owner, err := cfg.Db.GetOwnerById(req.Context(), effectiveOwnerID)
		if len(participantID) == 0 {
			participantID = owner.IdentificationNumber
		}
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		// Get gathering to check qualification rules
		gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering")
			return
		}

		// Parse qualification rules
		var unitTypes []string
		var floors []int64
		var entrances []int64

		if gathering.QualificationUnitTypes.Valid {
			json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
		}
		if gathering.QualificationFloors.Valid {
			json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
		}
		if gathering.QualificationEntrances.Valid {
			json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
		}

		// Get owner's units that meet qualification criteria
		ownerQualifiedUnits, err := cfg.Db.GetActiveOwnerUnitsForGathering(req.Context(), database.GetActiveOwnerUnitsForGatheringParams{
			AssociationID: int64(associationId),
			Column2:       len(unitTypes) > 0,
			UnitTypes:     unitTypes,
			Column4:       len(floors) > 0,
			UnitFloors:    floors,
			Column6:       len(entrances) > 0,
			UnitEntrances: entrances,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting owner qualified units", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get owner qualified units")
			return
		}

		// Filter units by this specific owner
		ownersUnits := make(map[int64]database.GetActiveOwnerUnitsForGatheringRow)
		for _, unit := range ownerQualifiedUnits {
			if unit.OwnerID == effectiveOwnerID {
				ownersUnits[unit.UnitID] = unit
			}
		}
		participationUnits := make([]int64, 0)
		totalPart := 0.0
		totalArea := 0.0
		for _, unitID := range addReq.UnitIDs {
			if _, ok := ownersUnits[unitID]; ok {
				slot, err := cfg.Db.AssignUnitSlot(req.Context(), database.AssignUnitSlotParams{
					GatheringID:   int64(gatheringId),
					UnitID:        unitID,
					ParticipantID: effectiveOwnerID,
				})
				if err != nil {
					logging.Logger.Log(zap.WarnLevel, "Error assigning unit slot", zap.Error(err))
					continue
				}
				if slot.ParticipantID != effectiveOwnerID {
					logging.Logger.Log(zap.WarnLevel, "Assigned unit slot to different participant", zap.Any("slot_participant_id", slot.ParticipantID), zap.Int64("effective_owner_id", effectiveOwnerID))
					continue
				}
				totalArea += ownersUnits[unitID].Area
				totalPart += ownersUnits[unitID].VotingWeight
				participationUnits = append(participationUnits, unitID)
			}
		}
		if len(participationUnits) == 0 {
			logging.Logger.Log(zap.WarnLevel, "No valid units found for participation", zap.Int64("owner_id", effectiveOwnerID))
			RespondWithError(rw, http.StatusBadRequest, "No valid units found for participation")
			return
		}
		participationUnitsBStr, err := json.Marshal(participationUnits)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error marshalling participation units", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to process participation units")
			return
		}

		participant, err := cfg.Db.CreateGatheringParticipant(req.Context(), database.CreateGatheringParticipantParams{
			GatheringID:               int64(gatheringId),
			ParticipantType:           addReq.ParticipantType,
			ParticipantName:           owner.Name,
			ParticipantIdentification: sql.NullString{String: participantID, Valid: true},
			OwnerID:                   sql.NullInt64{Int64: addReq.OwnerID, Valid: true},
			DelegatingOwnerID:         sql.NullInt64{Int64: delegatingOwnerID, Valid: addReq.DelegatingOwnerID != nil},
			DelegationDocumentRef:     sql.NullString{String: addReq.DelegationDocumentRef, Valid: addReq.DelegationDocumentRef != ""},
			UnitsInfo:                 string(participationUnitsBStr),
			UnitsPart:                 totalPart,
			UnitsArea:                 totalArea,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating gathering participation units", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering participant")
		}
		// Update gathering statistics
		go updateGatheringStats(cfg, int64(gatheringId), int64(associationId))

		RespondWithJSON(rw, http.StatusCreated, dbParticipantToResponse(participant))
	}
}

func HandleGetParticipants(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		participants, err := cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting participants", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get participants")
			return
		}

		// Get ballots to check who has voted
		ballots, _ := cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringId))
		votedMap := make(map[int64]bool)
		for _, b := range ballots {
			if b.IsValid.Bool {
				votedMap[b.ParticipantID] = true
			}
		}

		response := make([]GatheringParticipant, len(participants))
		for i, p := range participants {
			participant := dbParticipantRowToResponse(p)
			participant.HasVoted = votedMap[p.ID]
			response[i] = participant
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

func HandleCheckInParticipant(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		participantId, _ := strconv.Atoi(req.PathValue(ParticipantIdPathValue))

		err := cfg.Db.CheckInParticipant(req.Context(), database.CheckInParticipantParams{
			ID:          int64(participantId),
			GatheringID: int64(gatheringId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error checking in participant", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to check in participant")
			return
		}

		// Log audit
		cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringId),
			EntityType:  "participant",
			EntityID:    int64(participantId),
			Action:      "checked_in",
			PerformedBy: sql.NullString{String: req.Context().Value("userID").(string), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"time":"%s"}`, time.Now().Format(time.RFC3339)), Valid: true},
		})

		RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "checked_in"})
	}
}

func HandleSubmitBallot(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))
		participantId, _ := strconv.Atoi(req.PathValue(ParticipantIdPathValue))

		var ballotReq struct {
			BallotContent map[string]BallotVote `json:"ballot_content"`
		}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&ballotReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid ballot format")
			return
		}

		// Validate participant exists and hasn't voted
		_, err := cfg.Db.GetGatheringParticipant(req.Context(), database.GetGatheringParticipantParams{
			ID:          int64(participantId),
			GatheringID: int64(gatheringId),
		})
		if err != nil {
			RespondWithError(rw, http.StatusNotFound, "Participant not found")
			return
		}

		// Check if already voted
		existingBallot, _ := cfg.Db.GetBallotByParticipant(req.Context(), database.GetBallotByParticipantParams{
			GatheringID:   int64(gatheringId),
			ParticipantID: int64(participantId),
		})
		if existingBallot.ID > 0 && existingBallot.IsValid.Bool {
			RespondWithError(rw, http.StatusBadRequest, "Participant has already voted")
			return
		}

		// Create ballot JSON
		ballotJSON, err := json.Marshal(ballotReq.BallotContent)
		if err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid ballot content")
			return
		}

		// Calculate ballot hash
		hash := sha256.Sum256(ballotJSON)
		ballotHash := hex.EncodeToString(hash[:])

		// Submit ballot
		ballot, err := cfg.Db.CreateBallot(req.Context(), database.CreateBallotParams{
			GatheringID:        int64(gatheringId),
			ParticipantID:      int64(participantId),
			BallotContent:      string(ballotJSON),
			BallotHash:         ballotHash,
			SubmittedIp:        sql.NullString{String: req.RemoteAddr, Valid: true},
			SubmittedUserAgent: sql.NullString{String: req.UserAgent(), Valid: true},
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating ballot", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to submit ballot")
			return
		}

		// Update vote tallies asynchronously
		go updateVoteTallies(cfg, int64(gatheringId), participantId)

		// Log audit
		cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringId),
			EntityType:  "ballot",
			EntityID:    ballot.ID,
			Action:      "submitted",
			PerformedBy: sql.NullString{String: fmt.Sprintf("participant_%d", participantId), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"hash":"%s"}`, ballotHash), Valid: true},
		})

		RespondWithJSON(rw, http.StatusCreated, map[string]interface{}{
			"status":      "ballot_submitted",
			"ballot_hash": ballotHash,
			"ballot_id":   ballot.ID,
		})
	}
}

func HandleGetVoteResults(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		// Get gathering details
		gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Get all voting matters
		matters, err := cfg.Db.GetVotingMatters(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting voting matters", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting matters")
			return
		}

		// Get all ballots
		ballots, err := cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballots", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get ballots")
			return
		}

		// Get all participants to calculate totals
		participants, _ := cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringId))

		// Calculate participation stats
		votedParticipants := make(map[int64]bool)
		totalVotedWeight := 0.0
		totalVotedArea := 0.0
		for _, ballot := range ballots {
			if ballot.IsValid.Bool {
				votedParticipants[ballot.ParticipantID] = true
			}
		}

		for _, participant := range participants {

			if votedParticipants[participant.ID] {
				totalVotedWeight += participant.UnitsPart
				totalVotedArea += participant.UnitsArea
			}
		}

		// Build results
		results := VoteResults{
			GatheringID: int64(gatheringId),
			Results:     make([]VoteMatterResult, 0, len(matters)),
			Summary: GatheringSummary{
				QualifiedUnits:      int(gathering.QualifiedUnitsCount.Int64),
				QualifiedWeight:     gathering.QualifiedUnitsTotalPart.Float64,
				QualifiedArea:       gathering.QualifiedUnitsTotalArea.Float64,
				ParticipatingUnits:  int(gathering.ParticipatingUnitsCount.Int64),
				ParticipatingWeight: gathering.ParticipatingUnitsTotalPart.Float64,
				ParticipatingArea:   gathering.ParticipatingUnitsTotalArea.Float64,
				VotedUnits:          len(votedParticipants),
				VotedWeight:         totalVotedWeight,
				VotedArea:           totalVotedArea,
			},
		}

		// Calculate rates
		if results.Summary.QualifiedUnits > 0 {
			results.Summary.ParticipationRate = gathering.ParticipatingUnitsTotalArea.Float64 / gathering.QualifiedUnitsTotalArea.Float64 * 100
		}
		if results.Summary.ParticipatingUnits > 0 {
			results.Summary.VotingCompletionRate = totalVotedArea / gathering.QualifiedUnitsTotalArea.Float64 * 100
		}

		// Process each voting matter
		for _, matter := range matters {
			var votingConfig VotingConfig
			json.Unmarshal([]byte(matter.VotingConfig), &votingConfig)

			// Initialize tally for this matter
			tally := make(map[string]TallyResult)

			// Initialize options based on voting type
			if votingConfig.Type == "yes_no" {
				tally["yes"] = TallyResult{Count: 0, Weight: 0}
				tally["no"] = TallyResult{Count: 0, Weight: 0}
				if votingConfig.AllowAbstention {
					tally["abstain"] = TallyResult{Count: 0, Weight: 0}
				}
			} else if votingConfig.Type == "multiple_choice" {
				for _, option := range votingConfig.Options {
					tally[option.ID] = TallyResult{Count: 0, Weight: 0}
				}
				if votingConfig.AllowAbstention {
					tally["abstain"] = TallyResult{Count: 0, Weight: 0}
				}
			}

			// Count votes from valid ballots
			totalWeight := 0.0
			totalArea := 0.0
			for _, ballot := range ballots {
				if !ballot.IsValid.Bool {
					continue
				}

				// Get participant info for weight
				var participantWeight float64
				var participantArea float64
				for _, p := range participants {
					if p.ID == ballot.ParticipantID {

						participantWeight = p.UnitsPart
						participantArea = p.UnitsArea
						break
					}
				}

				// Parse ballot content
				var ballotContent map[string]BallotVote
				json.Unmarshal([]byte(ballot.BallotContent), &ballotContent)

				// Find vote for this matter
				matterIDStr := strconv.FormatInt(matter.ID, 10)
				if vote, ok := ballotContent[matterIDStr]; ok {
					voteKey := vote.VoteValue
					if vote.OptionID != "" {
						voteKey = vote.OptionID
					}

					if _, exists := tally[voteKey]; exists {
						tally[voteKey] = TallyResult{
							Count:  tally[voteKey].Count + 1,
							Weight: tally[voteKey].Weight + participantWeight,
							Area:   tally[voteKey].Area + participantArea,
						}
						if voteKey != "abstain" {
							totalWeight += participantWeight
							totalArea += participantArea
						}
					}
				}
			}

			// Calculate percentages
			for key, result := range tally {
				if totalWeight > 0 {
					result.Percentage = (result.Area / totalVotedArea) * 100
					tally[key] = result
				}
			}

			matterResult := VoteMatterResult{
				MatterID:       matter.ID,
				MatterTitle:    matter.Title,
				MatterType:     matter.MatterType,
				VotingConfig:   votingConfig,
				Tally:          tally,
				TotalVoted:     totalWeight,
				TotalAbstained: tally["abstain"].Weight,
			}

			matterResult.Passed = calculateIfPassed(matterResult, votingConfig)
			results.Results = append(results.Results, matterResult)
		}

		RespondWithJSON(rw, http.StatusOK, results)
	}
}

func HandleGetQualifiedUnits(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		// Get gathering to check qualification rules
		gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Parse qualification rules
		var unitTypes []string
		var floors []int64
		var entrances []int64

		if gathering.QualificationUnitTypes.Valid {
			json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
		}
		if gathering.QualificationFloors.Valid {
			json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
		}
		if gathering.QualificationEntrances.Valid {
			json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
		}

		// Get qualified units with owner information
		units, err := cfg.Db.GetActiveOwnerUnitsForGathering(req.Context(), database.GetActiveOwnerUnitsForGatheringParams{
			AssociationID: int64(associationId),
			Column2:       len(unitTypes) > 0,
			UnitTypes:     unitTypes,
			Column4:       len(floors) > 0,
			UnitFloors:    floors,
			Column6:       len(entrances) > 0,
			UnitEntrances: entrances,
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting qualified units", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get qualified units")
			return
		}

		type QualifiedUnit struct {
			ID              int64   `json:"id"`
			UnitNumber      string  `json:"unit_number"`
			CadastralNumber string  `json:"cadastral_number"`
			Floor           int     `json:"floor"`
			Entrance        int     `json:"entrance"`
			Area            float64 `json:"area"`
			Part            float64 `json:"part"`
			UnitType        string  `json:"unit_type"`
			BuildingName    string  `json:"building_name"`
			BuildingAddress string  `json:"building_address"`
			IsParticipating bool    `json:"is_participating"`
			OwnerID         int64   `json:"owner_id"`
			OwnerName       string  `json:"owner_name"`
		}

		// Get current participants to mark which units are already participating
		participants, _ := cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringId))
		participatingUnits := make(map[int64]bool)
		for _, p := range participants {
			partUnits := make([]int64, 0)
			json.Unmarshal([]byte(p.UnitsInfo), &partUnits)
			for _, unitID := range partUnits {
				participatingUnits[unitID] = true
			}
		}

		response := make([]QualifiedUnit, len(units))
		for i, u := range units {
			response[i] = QualifiedUnit{
				ID:              u.UnitID,
				UnitNumber:      u.UnitNumber,
				CadastralNumber: u.CadastralNumber,
				Floor:           int(u.Floor),
				Entrance:        int(u.Entrance),
				Area:            u.Area,
				Part:            u.VotingWeight,
				UnitType:        u.UnitType,
				BuildingName:    u.BuildingName,
				BuildingAddress: u.BuildingAddress,
				IsParticipating: participatingUnits[u.UnitID],
				OwnerID:         u.OwnerID,
				OwnerName:       u.OwnerName,
			}
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

func HandleGetNonParticipatingOwners(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		// Get gathering qualification rules
		gathering, err := cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Parse qualification rules
		var unitTypes []string
		var floors []int64
		var entrances []int64

		if gathering.QualificationUnitTypes.Valid {
			json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
		}
		if gathering.QualificationFloors.Valid {
			json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
		}
		if gathering.QualificationEntrances.Valid {
			json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
		}

		owners, err := cfg.Db.GetNonParticipatingOwners(req.Context(), database.GetNonParticipatingOwnersParams{
			AssociationID:   int64(associationId),
			AssociationID_2: int64(associationId),
			Column3:         len(unitTypes) > 0,
			UnitTypes:       unitTypes,
			Column5:         len(floors) > 0,
			UnitFloors:      floors,
			Column7:         len(entrances) > 0,
			UnitEntrances:   entrances,
			GatheringID:     int64(gatheringId),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting non-participating owners", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get non-participating owners")
			return
		}

		type NonParticipatingOwner struct {
			ID                   int64  `json:"id"`
			Name                 string `json:"name"`
			IdentificationNumber string `json:"identification_number"`
			ContactEmail         string `json:"contact_email"`
			ContactPhone         string `json:"contact_phone"`
			UnitsCount           int    `json:"units_count"`
		}

		response := make([]NonParticipatingOwner, len(owners))
		for i, o := range owners {
			response[i] = NonParticipatingOwner{
				ID:                   o.ID,
				Name:                 o.Name,
				IdentificationNumber: o.IdentificationNumber,
				ContactEmail:         o.ContactEmail,
				ContactPhone:         o.ContactPhone,
				UnitsCount:           int(o.UnitsCount),
			}
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

func HandleGetGatheringStats(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationId, _ := strconv.Atoi(req.PathValue(AssociationIdPathValue))
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		// Get gathering details
		gathering, err := cfg.Db.GetGatheringStats(req.Context(), database.GetGatheringStatsParams{
			ID:            int64(gatheringId),
			AssociationID: int64(associationId),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering stats")
			return
		}

		// Get participant count
		participantCount, err := cfg.Db.GetGatheringParticipantCount(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting participant count", zap.Error(err))
		}

		// Get ballot count
		ballotCount, err := cfg.Db.GetGatheringBallotCount(req.Context(), int64(gatheringId))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballot count", zap.Error(err))
		}

		// Calculate voted weight by getting all participants who voted
		participants, _ := cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringId))
		ballots, _ := cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringId))

		// Create a map of participant IDs who voted
		votedParticipants := make(map[int64]bool)
		for _, ballot := range ballots {
			if ballot.IsValid.Bool {
				votedParticipants[ballot.ParticipantID] = true
			}
		}

		// Calculate total voted weight
		totalVotedWeight := 0.0
		totalVotedArea := 0.0
		for _, participant := range participants {
			if votedParticipants[participant.ID] {
				totalVotedWeight += participant.UnitsPart
				totalVotedArea += participant.UnitsArea
			}
		}

		type GatheringStats struct {
			Gathering
			ParticipantCount        int     `json:"participant_count"`
			VotedCount              int     `json:"voted_count"`
			TotalVotedWeight        float64 `json:"total_voted_weight"`
			TotalVotedArea          float64 `json:"total_voted_area"`
			ParticipationPercentage float64 `json:"participation_percentage"`
			VotingPercentage        float64 `json:"voting_percentage"`
		}

		response := GatheringStats{
			Gathering:        dbGatheringToResponse(gathering),
			ParticipantCount: int(participantCount),
			VotedCount:       int(ballotCount),
			TotalVotedWeight: totalVotedWeight,
			TotalVotedArea:   totalVotedArea,
		}

		// Calculate percentages
		// review to catter for multiunit participants
		if gathering.QualifiedUnitsCount.Int64 > 0 {
			response.ParticipationPercentage = float64(participantCount) / float64(gathering.QualifiedUnitsCount.Int64) * 100
		}
		if participantCount > 0 {
			response.VotingPercentage = float64(ballotCount) / float64(participantCount) * 100
		}

		RespondWithJSON(rw, http.StatusOK, response)
	}
}

// Helper functions

func dbGatheringToResponse(g database.Gathering) Gathering {
	var unitTypes []string
	var floors []int64
	var entrances []int64

	if g.QualificationUnitTypes.Valid {
		json.Unmarshal([]byte(g.QualificationUnitTypes.String), &unitTypes)
	}
	if g.QualificationFloors.Valid {
		json.Unmarshal([]byte(g.QualificationFloors.String), &floors)
	}
	if g.QualificationEntrances.Valid {
		json.Unmarshal([]byte(g.QualificationEntrances.String), &entrances)
	}

	return Gathering{
		ID:                          g.ID,
		AssociationID:               g.AssociationID,
		Title:                       g.Title,
		Description:                 g.Description,
		Intent:                      g.Intent,
		GatheringDate:               g.GatheringDate,
		GatheringType:               g.GatheringType,
		Status:                      g.Status,
		QualificationUnitTypes:      unitTypes,
		QualificationFloors:         floors,
		QualificationEntrances:      entrances,
		QualificationCustomRule:     g.QualificationCustomRule.String,
		QualifiedUnitsCount:         int(g.QualifiedUnitsCount.Int64),
		QualifiedUnitsTotalPart:     g.QualifiedUnitsTotalPart.Float64,
		QualifiedUnitsTotalArea:     g.QualifiedUnitsTotalArea.Float64,
		ParticipatingUnitsCount:     int(g.ParticipatingUnitsCount.Int64),
		ParticipatingUnitsTotalPart: g.ParticipatingUnitsTotalPart.Float64,
		ParticipatingUnitsTotalArea: g.ParticipatingUnitsTotalArea.Float64,
		CreatedAt:                   g.CreatedAt.Time,
		UpdatedAt:                   g.UpdatedAt.Time,
	}
}

func dbVotingMatterToResponse(m database.VotingMatter) VotingMatter {
	var config VotingConfig
	json.Unmarshal([]byte(m.VotingConfig), &config)

	return VotingMatter{
		ID:           m.ID,
		GatheringID:  m.GatheringID,
		OrderIndex:   int(m.OrderIndex),
		Title:        m.Title,
		Description:  m.Description.String,
		MatterType:   m.MatterType,
		VotingConfig: config,
		CreatedAt:    m.CreatedAt.Time,
		UpdatedAt:    m.UpdatedAt.Time,
	}
}

func dbParticipantToResponse(p database.GatheringParticipant) GatheringParticipant {
	var unitsInfo []int64
	json.Unmarshal([]byte(p.UnitsInfo), &unitsInfo)

	return GatheringParticipant{
		ID:                        p.ID,
		GatheringID:               p.GatheringID,
		ParticipantType:           p.ParticipantType,
		ParticipantName:           p.ParticipantName,
		ParticipantIdentification: p.ParticipantIdentification.String,
		OwnerID:                   nullInt64ToPtr(p.OwnerID),
		DelegatingOwnerID:         nullInt64ToPtr(p.DelegatingOwnerID),
		DelegationDocumentRef:     p.DelegationDocumentRef.String,
		UnitsInfo:                 unitsInfo,
		UnitsPart:                 p.UnitsPart,
		UnitsArea:                 p.UnitsArea,
		CheckInTime:               nullTimeToPtr(p.CheckInTime),
		CreatedAt:                 p.CreatedAt.Time,
		UpdatedAt:                 p.UpdatedAt.Time,
	}
}

func dbParticipantRowToResponse(p database.GetGatheringParticipantsRow) GatheringParticipant {
	var unitsInfo []int64
	json.Unmarshal([]byte(p.UnitsInfo), &unitsInfo)

	return GatheringParticipant{
		ID:                        p.ID,
		GatheringID:               p.GatheringID,
		ParticipantType:           p.ParticipantType,
		ParticipantName:           p.ParticipantName,
		ParticipantIdentification: p.ParticipantIdentification.String,
		OwnerID:                   nullInt64ToPtr(p.OwnerID),
		DelegatingOwnerID:         nullInt64ToPtr(p.DelegatingOwnerID),
		DelegationDocumentRef:     p.DelegationDocumentRef.String,
		UnitsInfo:                 unitsInfo,
		UnitsPart:                 p.UnitsPart,
		UnitsArea:                 p.UnitsArea,
		CheckInTime:               nullTimeToPtr(p.CheckInTime),
		CreatedAt:                 p.CreatedAt.Time,
		UpdatedAt:                 p.UpdatedAt.Time,
	}
}

func nullInt64ToPtr(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

func nullTimeToPtr(n sql.NullTime) *time.Time {
	if n.Valid {
		return &n.Time
	}
	return nil
}

func calculateIfPassed(result VoteMatterResult, config VotingConfig) bool {
	if result.TotalVoted == 0 {
		return false
	}

	// Check quorum first
	totalPossibleWeight := result.TotalVoted + result.TotalAbstained
	if config.Quorum > 0 && totalPossibleWeight > 0 {
		quorumPercentage := (result.TotalVoted / totalPossibleWeight) * 100
		if quorumPercentage < config.Quorum {
			return false
		}
	}

	// For yes/no votes
	if config.Type == "yes_no" {
		yesVotes := result.Tally["yes"].Weight
		totalValidVotes := result.TotalVoted
		if !config.AllowAbstention && result.TotalAbstained > 0 {
			totalValidVotes += result.TotalAbstained
		}

		percentage := (yesVotes / totalValidVotes) * 100

		switch config.RequiredMajority {
		case "simple":
			return percentage > 50
		case "supermajority":
			return percentage >= 66.67
		case "custom":
			return percentage >= config.RequiredMajorityValue
		}
	}

	// For multiple choice, the option with most votes wins if it meets the threshold
	if config.Type == "multiple_choice" {
		var maxWeight float64
		for _, tally := range result.Tally {
			if tally.Weight > maxWeight {
				maxWeight = tally.Weight
			}
		}

		percentage := (maxWeight / result.TotalVoted) * 100

		switch config.RequiredMajority {
		case "simple":
			return percentage > 50
		case "supermajority":
			return percentage >= 66.67
		case "custom":
			return percentage >= config.RequiredMajorityValue
		}
	}

	return false
}

func updateGatheringStats(cfg *ApiConfig, gatheringID, associationID int64) (int, float64, float64) {
	ctx := context.Background()

	// Get gathering details
	gathering, err := cfg.Db.GetGathering(ctx, database.GetGatheringParams{
		ID:            gatheringID,
		AssociationID: associationID,
	})
	if err != nil {
		return 0, 0.0, 0.0
	}

	// Parse qualification rules
	var unitTypes []string
	var floors []int64
	var entrances []int64

	if gathering.QualificationUnitTypes.Valid {
		json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
	}
	if gathering.QualificationFloors.Valid {
		json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
	}
	if gathering.QualificationEntrances.Valid {
		json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
	}

	// Get qualified units
	units, err := cfg.Db.GetQualifiedUnits(ctx, database.GetQualifiedUnitsParams{
		AssociationID: associationID,
		Column2:       len(unitTypes) > 0,
		UnitTypes:     unitTypes,
		Column4:       len(floors) > 0,
		UnitFloors:    floors,
		Column6:       len(entrances) > 0,
		UnitEntrances: entrances,
	})

	if err != nil {
		return 0, 0.0, 0.0
	}

	// Calculate totals
	qualifiedCount := len(units)
	qualifiedTotalPart := 0.0
	qualifiedTotalArea := 0.0
	for _, u := range units {
		qualifiedTotalPart += u.Part
		qualifiedTotalArea += u.Area
	}

	// Update stats
	cfg.Db.UpdateGatheringStats(ctx, database.UpdateGatheringStatsParams{
		QualifiedUnitsCount:     sql.NullInt64{Int64: int64(qualifiedCount), Valid: true},
		QualifiedUnitsTotalPart: sql.NullFloat64{Float64: qualifiedTotalPart, Valid: true},
		QualifiedUnitsTotalArea: sql.NullFloat64{Float64: qualifiedTotalArea, Valid: true},
		ID:                      gatheringID,
	})
	return qualifiedCount, qualifiedTotalPart, qualifiedTotalArea
}

func updateGatheringParticipationStats(cfg *ApiConfig, gatheringID, associationID int64) {
	ctx := context.Background()

	// Get gathering details
	gathering, err := cfg.Db.GetGathering(ctx, database.GetGatheringParams{
		ID:            gatheringID,
		AssociationID: associationID,
	})
	if err != nil {
		return
	}

	// Parse qualification rules
	var unitTypes []string
	var floors []int64
	var entrances []int64

	if gathering.QualificationUnitTypes.Valid {
		json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
	}
	if gathering.QualificationFloors.Valid {
		json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
	}
	if gathering.QualificationEntrances.Valid {
		json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
	}
	// Get participants
	participants, _ := cfg.Db.GetGatheringParticipants(ctx, gatheringID)
	participatingCount := len(participants)
	participatingTotalPart := 0.0
	participatingTotalArea := 0.0
	for _, p := range participants {
		participatingTotalPart += p.UnitsPart
		participatingTotalArea += p.UnitsArea
	}

	// Update stats
	cfg.Db.UpdateParticipationStats(ctx, database.UpdateParticipationStatsParams{
		ParticipatingUnitsCount:     sql.NullInt64{Int64: int64(participatingCount), Valid: true},
		ParticipatingUnitsTotalPart: sql.NullFloat64{Float64: participatingTotalPart, Valid: true},
		ParticipatingUnitsTotalArea: sql.NullFloat64{Float64: participatingTotalArea, Valid: true},
		ID:                          gatheringID,
	})
}

func updateVoteTallies(cfg *ApiConfig, gatheringID int64, participantId int) {

	logging.Logger.Log(zap.InfoLevel, "Vote tallies marked for update",
		zap.Int64("gathering_id", gatheringID))
}

func calculateFinalResults(cfg *ApiConfig, gatheringID int64) {
	// This function would be called when gathering is closed
	// It would finalize all tallies and potentially generate reports
	updateVoteTallies(cfg, gatheringID, -1)
}

// Additional handlers for notifications and audit logs

func HandleSendNotification(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		var notifyReq struct {
			NotificationType string  `json:"notification_type"` // invitation, reminder, results
			OwnerIDs         []int64 `json:"owner_ids,omitempty"`
			SendVia          string  `json:"send_via"` // email, sms
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&notifyReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// If no specific owners, get all non-participating owners
		if len(notifyReq.OwnerIDs) == 0 && notifyReq.NotificationType == "invitation" {
			// Get gathering and qualification rules...
			// Then get non-participating owners...
		}

		// Send notifications
		var sent []database.VotingNotification
		for _, ownerID := range notifyReq.OwnerIDs {
			notification, err := cfg.Db.CreateNotification(req.Context(), database.CreateNotificationParams{
				GatheringID:      int64(gatheringId),
				OwnerID:          ownerID,
				NotificationType: notifyReq.NotificationType,
				SentVia:          sql.NullString{String: notifyReq.SendVia, Valid: true},
			})

			if err == nil {
				sent = append(sent, notification)
				// Here you would actually send the notification via email/SMS
			}
		}

		RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"sent_count":    len(sent),
			"notifications": sent,
		})
	}
}

func HandleGetAuditLogs(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringId, _ := strconv.Atoi(req.PathValue(GatheringIdPathValue))

		limitStr := req.URL.Query().Get("limit")
		limit := 100
		if limitStr != "" {
			if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
				limit = l
			}
		}

		logs, err := cfg.Db.GetAuditLogs(req.Context(), database.GetAuditLogsParams{
			GatheringID: int64(gatheringId),
			Limit:       int64(limit),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting audit logs", zap.Error(err))
			RespondWithError(rw, http.StatusInternalServerError, "Failed to get audit logs")
			return
		}

		RespondWithJSON(rw, http.StatusOK, logs)
	}
}

func HandleVerifyBallot(cfg *ApiConfig) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var verifyReq struct {
			BallotID   int64  `json:"ballot_id"`
			BallotHash string `json:"ballot_hash"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&verifyReq); err != nil {
			RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Get ballot from database
		ballots, err := cfg.Db.GetBallotsForGathering(req.Context(), verifyReq.BallotID) // You'd need a GetBallotByID query
		if err != nil || len(ballots) == 0 {
			RespondWithError(rw, http.StatusNotFound, "Ballot not found")
			return
		}

		// Find the specific ballot
		var ballot database.VotingBallot
		found := false
		for _, b := range ballots {
			if b.ID == verifyReq.BallotID {
				ballot = ballotRowToVotingBallot(b)
				found = true
				break
			}
		}

		if !found {
			RespondWithError(rw, http.StatusNotFound, "Ballot not found")
			return
		}

		// Verify hash
		hash := sha256.Sum256([]byte(ballot.BallotContent))
		calculatedHash := hex.EncodeToString(hash[:])

		RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"valid":        calculatedHash == verifyReq.BallotHash && ballot.BallotHash == verifyReq.BallotHash,
			"ballot_id":    ballot.ID,
			"submitted_at": ballot.SubmittedAt,
			"is_valid":     ballot.IsValid,
		})
	}
}

func ballotRowToVotingBallot(b database.GetBallotsForGatheringRow) database.VotingBallot {
	return database.VotingBallot{
		ID:                   b.ID,
		GatheringID:          b.GatheringID,
		ParticipantID:        b.ParticipantID,
		BallotContent:        b.BallotContent,
		BallotHash:           b.BallotHash,
		SubmittedAt:          b.SubmittedAt,
		SubmittedIp:          b.SubmittedIp,
		SubmittedUserAgent:   b.SubmittedUserAgent,
		IsValid:              b.IsValid,
		SignatureCertificate: b.SignatureCertificate,
		SignatureTimestamp:   b.SignatureTimestamp,
		Signature:            b.Signature,
		InvalidationReason:   b.InvalidationReason,
		InvalidatedAt:        b.InvalidatedAt,
	}
}
