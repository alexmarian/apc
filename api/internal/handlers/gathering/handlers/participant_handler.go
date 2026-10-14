package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/services"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// ParticipantHandler handles participant operations
type ParticipantHandler struct {
	cfg              *handlers.ApiConfig
	gatheringHandler *GatheringHandler
	statsService     *services.StatsService
}

// NewParticipantHandler creates a new ParticipantHandler
func NewParticipantHandler(cfg *handlers.ApiConfig, gatheringHandler *GatheringHandler) *ParticipantHandler {
	return &ParticipantHandler{
		cfg:              cfg,
		gatheringHandler: gatheringHandler,
		statsService:     services.NewStatsService(cfg.Db),
	}
}

// HandleGetParticipants returns all participants for a gathering
func (h *ParticipantHandler) HandleGetParticipants() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		participants, err := h.cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting participants", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get participants")
			return
		}

		// Get ballots to check who has voted
		ballots, _ := h.cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringID))
		votedMap := make(map[int64]bool)
		for _, b := range ballots {
			if b.IsValid.Bool {
				votedMap[b.ParticipantID] = true
			}
		}

		associationID, _ := strconv.ParseInt(req.PathValue(handlers.AssociationIdPathValue), 10, 64)
		photos := handlers.OwnerPhotoURLs(req.Context(), h.cfg, associationID)
		response := make([]domain.GatheringParticipant, len(participants))
		for i, p := range participants {
			participant := domain.DBParticipantRowToResponse(p)
			participant.HasVoted = votedMap[p.ID]
			if participant.OwnerID != nil {
				participant.PhotoURL = photos[*participant.OwnerID]
			}
			response[i] = participant
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleAddParticipant adds a participant to a gathering
func (h *ParticipantHandler) HandleAddParticipant() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		active := h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "active")
		if !active {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Cannot add a participant in a non-active gathering")
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
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Validate that unit IDs are provided and valid
		if len(addReq.UnitIDs) == 0 {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Unit IDs are required for participation")
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
			handlers.RespondWithError(rw, http.StatusBadRequest, "At least one valid unit ID is required for participation")
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
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid participant type or owner ID")
			return
		}

		owner, err := h.cfg.Db.GetOwnerById(req.Context(), effectiveOwnerID)
		if len(participantID) == 0 {
			participantID = owner.IdentificationNumber
		}
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		// Get gathering to check qualification rules
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting gathering", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get gathering")
			return
		}

		// Get owner's units that meet qualification criteria
		ownerQualifiedUnits, err := h.cfg.Db.GetActiveOwnerUnitsForGathering(req.Context(), domain.QualificationOf(gathering).OwnerUnitsParams(int64(associationID)))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting owner qualified units", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get owner qualified units")
			return
		}

		// Filter units by this specific owner
		ownersUnits := make(map[int64]database.GetActiveOwnerUnitsForGatheringRow)
		for _, unit := range ownerQualifiedUnits {
			if unit.OwnerID == effectiveOwnerID {
				ownersUnits[unit.UnitID] = unit
			}
		}

		// The participant first, then its units' slots under its own id (they
		// were claimed with the owner id as participant id before, and before
		// the participant existed). Units already taken are skipped; with none
		// left nothing is kept.
		tx, err := h.cfg.DB.BeginTx(req.Context(), nil)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering participant")
			return
		}
		defer tx.Rollback()
		qtx := h.cfg.Db.WithTx(tx)

		participant, err := qtx.CreateGatheringParticipant(req.Context(), database.CreateGatheringParticipantParams{
			GatheringID:               int64(gatheringID),
			ParticipantType:           addReq.ParticipantType,
			ParticipantName:           owner.Name,
			ParticipantIdentification: sql.NullString{String: participantID, Valid: true},
			OwnerID:                   sql.NullInt64{Int64: addReq.OwnerID, Valid: addReq.ParticipantType == "owner"},
			DelegatingOwnerID:         sql.NullInt64{Int64: delegatingOwnerID, Valid: addReq.DelegatingOwnerID != nil},
			DelegationDocumentRef:     sql.NullString{String: addReq.DelegationDocumentRef, Valid: addReq.DelegationDocumentRef != ""},
			UnitsInfo:                 "[]",
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating gathering participant", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering participant")
			return
		}

		participationUnits := make([]int64, 0)
		totalPart := 0.0
		totalArea := 0.0
		for _, unitID := range addReq.UnitIDs {
			unit, ok := ownersUnits[unitID]
			if !ok {
				continue
			}
			_, err := qtx.AssignUnitSlot(req.Context(), database.AssignUnitSlotParams{
				GatheringID:   int64(gatheringID),
				UnitID:        unitID,
				ParticipantID: participant.ID,
			})
			if errors.Is(err, sql.ErrNoRows) {
				continue // already taken
			}
			if err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error assigning unit slot", zap.Error(err))
				handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering participant")
				return
			}
			totalArea += unit.Area
			totalPart += unit.VotingWeight
			participationUnits = append(participationUnits, unitID)
		}
		if len(participationUnits) == 0 {
			handlers.RespondWithError(rw, http.StatusBadRequest, "No valid units found for participation")
			return
		}
		unitsJSON, _ := json.Marshal(participationUnits)
		participant, err = qtx.UpdateGatheringParticipantsUnits(req.Context(), database.UpdateGatheringParticipantsUnitsParams{
			UnitsInfo:   string(unitsJSON),
			UnitsArea:   totalArea,
			UnitsPart:   totalPart,
			ID:          participant.ID,
			GatheringID: int64(gatheringID),
		})
		if err != nil || tx.Commit() != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create gathering participant")
			return
		}

		// Update gathering statistics
		h.statsService.UpdateGatheringParticipationStats(int64(gatheringID), int64(associationID))

		handlers.RespondWithJSON(rw, http.StatusCreated, domain.DBParticipantToResponse(participant))
	}
}

// HandleCheckInParticipant checks in a participant
func (h *ParticipantHandler) HandleCheckInParticipant() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))
		participantID, _ := strconv.Atoi(req.PathValue(domain.ParticipantIDPathValue))

		// Participants can't change once the gathering is tallied.
		if h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "tallied") {
			handlers.RespondWithError(rw, http.StatusConflict, "The gathering is tallied")
			return
		}

		checkedIn, err := h.cfg.Db.CheckInParticipant(req.Context(), database.CheckInParticipantParams{
			ID:          int64(participantID),
			GatheringID: int64(gatheringID),
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error checking in participant", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to check in participant")
			return
		}
		if checkedIn == 0 {
			handlers.RespondWithError(rw, http.StatusNotFound, "Participant not found")
			return
		}

		// Log audit
		h.cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringID),
			EntityType:  "participant",
			EntityID:    int64(participantID),
			Action:      "checked_in",
			PerformedBy: sql.NullString{String: handlers.GetUserIdFromContext(req), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"time":"%s"}`, time.Now().Format(time.RFC3339)), Valid: true},
		})

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]string{"status": "checked_in"})
	}
}
