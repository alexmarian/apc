package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
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

// BallotHandler handles ballot and voting operations
type BallotHandler struct {
	cfg              *handlers.ApiConfig
	gatheringHandler *GatheringHandler
	statsService     *services.StatsService
	tallyService     *services.TallyService
}

// NewBallotHandler creates a new BallotHandler
func NewBallotHandler(cfg *handlers.ApiConfig, gatheringHandler *GatheringHandler) *BallotHandler {
	return &BallotHandler{
		cfg:              cfg,
		gatheringHandler: gatheringHandler,
		statsService:     services.NewStatsService(cfg.Db),
		tallyService:     services.NewTallyService(cfg.Db),
	}
}

// HandleSubmitBallot handles ballot submission
func (h *BallotHandler) HandleSubmitBallot() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		var ballotReq struct {
			VoterType             string                       `json:"voter_type"`
			OwnerID               int64                        `json:"owner_id"`
			DelegatingOwnerID     *int64                       `json:"delegating_owner_id,omitempty"`
			DelegationDocumentRef string                       `json:"delegation_document_ref,omitempty"`
			UnitIDs               []int64                      `json:"unit_ids"`
			BallotContent         map[string]domain.BallotVote `json:"ballot_content"`
		}
		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&ballotReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid ballot format")
			return
		}

		// Validate gathering is in active state
		active := h.gatheringHandler.ValidateGatheringStateWithFetch(req, gatheringID, associationID, "active")
		if !active {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Gathering is not active")
			return
		}

		// Validate required fields
		if len(ballotReq.UnitIDs) == 0 {
			handlers.RespondWithError(rw, http.StatusBadRequest, "At least one unit ID is required")
			return
		}

		// Determine the effective owner ID
		var effectiveOwnerID int64
		if ballotReq.VoterType == "owner" {
			effectiveOwnerID = ballotReq.OwnerID
		} else if ballotReq.VoterType == "delegate" && ballotReq.DelegatingOwnerID != nil {
			effectiveOwnerID = *ballotReq.DelegatingOwnerID
		} else {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid voter type or missing owner information")
			return
		}

		// Get owner information
		owner, err := h.cfg.Db.GetOwnerById(req.Context(), effectiveOwnerID)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		// Get eligible voters to validate units and calculate weights
		eligibleRows, err := h.cfg.Db.GetEligibleVotersWithUnits(req.Context(), database.GetEligibleVotersWithUnitsParams{
			GatheringID:   int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting eligible voters", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to validate units")
			return
		}

		// Build map of units owned by this owner
		ownerUnits := make(map[int64]database.GetEligibleVotersWithUnitsRow)
		for _, row := range eligibleRows {
			if row.OwnerID == effectiveOwnerID {
				ownerUnits[row.UnitID] = row
			}
		}

		// Validate all requested units are owned by this owner and available
		totalArea := 0.0
		totalWeight := 0.0
		validUnitIDs := make([]int64, 0)
		for _, unitID := range ballotReq.UnitIDs {
			unit, exists := ownerUnits[unitID]
			if !exists {
				handlers.RespondWithError(rw, http.StatusBadRequest, fmt.Sprintf("Unit %d is not owned by this owner or not qualified", unitID))
				return
			}
			if unit.IsAvailable == 0 {
				handlers.RespondWithError(rw, http.StatusBadRequest, fmt.Sprintf("Unit %d is not available (already assigned)", unitID))
				return
			}
			validUnitIDs = append(validUnitIDs, unitID)
			totalArea += unit.Area
			totalWeight += unit.VotingWeight
		}

		// Get or create participant
		participantID := sql.NullString{String: owner.IdentificationNumber, Valid: true}
		if ballotReq.VoterType == "delegate" {
			participantID = sql.NullString{String: ballotReq.DelegationDocumentRef, Valid: true}
		}

		unitsJSON, _ := json.Marshal(validUnitIDs)

		participant, err := h.cfg.Db.CreateGatheringParticipant(req.Context(), database.CreateGatheringParticipantParams{
			GatheringID:               int64(gatheringID),
			ParticipantType:           ballotReq.VoterType,
			ParticipantName:           owner.Name,
			ParticipantIdentification: participantID,
			OwnerID:                   sql.NullInt64{Int64: ballotReq.OwnerID, Valid: ballotReq.VoterType == "owner"},
			DelegatingOwnerID:         sql.NullInt64{Int64: effectiveOwnerID, Valid: ballotReq.VoterType == "delegate"},
			DelegationDocumentRef:     sql.NullString{String: ballotReq.DelegationDocumentRef, Valid: ballotReq.VoterType == "delegate"},
			UnitsInfo:                 string(unitsJSON),
			UnitsArea:                 totalArea,
			UnitsPart:                 totalWeight,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating participant", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to create participant")
			return
		}

		// Assign unit slots to this participant
		for _, unitID := range validUnitIDs {
			_, err := h.cfg.Db.AssignUnitSlot(req.Context(), database.AssignUnitSlotParams{
				ParticipantID: participant.ID,
				GatheringID:   int64(gatheringID),
				UnitID:        unitID,
			})
			if err != nil {
				logging.Logger.Log(zap.WarnLevel, "Error assigning unit slot", zap.Error(err), zap.Int64("unit_id", unitID))
			}
		}

		// Create ballot JSON
		ballotJSON, err := json.Marshal(ballotReq.BallotContent)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid ballot content")
			return
		}

		// Calculate ballot hash
		hash := sha256.Sum256(ballotJSON)
		ballotHash := hex.EncodeToString(hash[:])

		// Submit ballot
		ballot, err := h.cfg.Db.CreateBallot(req.Context(), database.CreateBallotParams{
			GatheringID:        int64(gatheringID),
			ParticipantID:      participant.ID,
			BallotContent:      string(ballotJSON),
			BallotHash:         ballotHash,
			SubmittedIp:        sql.NullString{String: req.RemoteAddr, Valid: true},
			SubmittedUserAgent: sql.NullString{String: req.UserAgent(), Valid: true},
		})

		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating ballot", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to submit ballot")
			return
		}

		// Update gathering stats
		go h.statsService.UpdateGatheringParticipationStats(int64(gatheringID), int64(associationID))

		// Update vote tallies asynchronously
		go h.tallyService.UpdateVoteTallies(int64(gatheringID), int(participant.ID))

		// Log audit
		h.cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringID),
			EntityType:  "ballot",
			EntityID:    ballot.ID,
			Action:      "submitted",
			PerformedBy: sql.NullString{String: fmt.Sprintf("participant_%d", participant.ID), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"hash":"%s","voter_type":"%s"}`, ballotHash, ballotReq.VoterType), Valid: true},
		})

		handlers.RespondWithJSON(rw, http.StatusCreated, map[string]interface{}{
			"status":         "ballot_submitted",
			"ballot_hash":    ballotHash,
			"ballot_id":      ballot.ID,
			"participant_id": participant.ID,
		})
	}
}

// HandleGetBallots returns all ballots for a gathering
func (h *BallotHandler) HandleGetBallots() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get all ballots
		ballots, err := h.cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballots", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get ballots")
			return
		}

		// Build response with metadata only (no ballot content)
		type BallotMetadata struct {
			ID              int64   `json:"id"`
			ParticipantName string  `json:"participant_name"`
			UnitsInfo       string  `json:"units_info"`
			UnitsArea       float64 `json:"units_area"`
			UnitsPart       float64 `json:"units_part"`
			BallotHash      string  `json:"ballot_hash"`
			SubmittedAt     string  `json:"submitted_at"`
			IsValid         bool    `json:"is_valid"`
		}

		response := make([]BallotMetadata, len(ballots))
		for i, b := range ballots {
			submittedAt := ""
			if b.SubmittedAt.Valid {
				submittedAt = b.SubmittedAt.Time.Format(time.RFC3339)
			}

			response[i] = BallotMetadata{
				ID:              b.ID,
				ParticipantName: b.ParticipantName,
				UnitsInfo:       b.UnitsInfo,
				UnitsArea:       b.UnitsArea,
				UnitsPart:       b.UnitsPart,
				BallotHash:      b.BallotHash,
				SubmittedAt:     submittedAt,
				IsValid:         b.IsValid.Bool,
			}
		}

		handlers.RespondWithJSON(rw, http.StatusOK, response)
	}
}

// HandleVerifyBallot verifies a ballot's integrity
func (h *BallotHandler) HandleVerifyBallot() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		var verifyReq struct {
			BallotID   int64  `json:"ballot_id"`
			BallotHash string `json:"ballot_hash"`
		}

		decoder := json.NewDecoder(req.Body)
		if err := decoder.Decode(&verifyReq); err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Invalid request format")
			return
		}

		// Get ballot from database
		ballots, err := h.cfg.Db.GetBallotsForGathering(req.Context(), verifyReq.BallotID)
		if err != nil || len(ballots) == 0 {
			handlers.RespondWithError(rw, http.StatusNotFound, "Ballot not found")
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
			handlers.RespondWithError(rw, http.StatusNotFound, "Ballot not found")
			return
		}

		// Verify hash
		hash := sha256.Sum256([]byte(ballot.BallotContent))
		calculatedHash := hex.EncodeToString(hash[:])

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"valid":        calculatedHash == verifyReq.BallotHash && ballot.BallotHash == verifyReq.BallotHash,
			"ballot_id":    ballot.ID,
			"submitted_at": ballot.SubmittedAt,
			"is_valid":     ballot.IsValid,
		})
	}
}

// Helper function to convert ballot row to voting ballot
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
