package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	cfg                  *handlers.ApiConfig
	gatheringHandler     *GatheringHandler
	statsService         *services.StatsService
	tallyService         *services.TallyService
	votingResultsService *services.VotingResultsService
}

// NewBallotHandler creates a new BallotHandler
func NewBallotHandler(cfg *handlers.ApiConfig, gatheringHandler *GatheringHandler) *BallotHandler {
	tallyService := services.NewTallyService(cfg.Db)
	quorumService := services.NewQuorumService(cfg.Db)
	return &BallotHandler{
		cfg:                  cfg,
		gatheringHandler:     gatheringHandler,
		statsService:         services.NewStatsService(cfg.Db),
		tallyService:         tallyService,
		votingResultsService: services.NewVotingResultsService(cfg.Db, quorumService, tallyService),
	}
}

// HandleSubmitBallot records a ballot cast on behalf of an owner, in person
// (voter_type "owner") or through a delegate ("delegate": delegate_name and
// delegate_identification name who voted; without them the owner's name and
// the delegation document are recorded, as before). The units must be the
// owner's, qualified and not yet voted; the content must answer every
// voting matter. All of it is recorded atomically (see recordBallot).
func (h *BallotHandler) HandleSubmitBallot() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		var ballotReq struct {
			VoterType              string                       `json:"voter_type"`
			OwnerID                int64                        `json:"owner_id"`
			DelegatingOwnerID      *int64                       `json:"delegating_owner_id,omitempty"`
			DelegationDocumentRef  string                       `json:"delegation_document_ref,omitempty"`
			DelegateName           string                       `json:"delegate_name,omitempty"`
			DelegateIdentification string                       `json:"delegate_identification,omitempty"`
			UnitIDs                []int64                      `json:"unit_ids"`
			BallotContent          map[string]domain.BallotVote `json:"ballot_content"`
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

		owner, err := h.cfg.Db.GetOwnerById(req.Context(), effectiveOwnerID)
		if err != nil || owner.AssociationID != int64(associationID) {
			handlers.RespondWithError(rw, http.StatusBadRequest, "Owner not found")
			return
		}

		matters, err := h.cfg.Db.GetVotingMatters(req.Context(), int64(gatheringID))
		if err != nil {
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to load voting matters")
			return
		}
		content, err := validateBallotContent(matters, ballotReq.BallotContent)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusBadRequest, err.Error())
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
		units := make([]database.GetEligibleVotersWithUnitsRow, 0, len(ballotReq.UnitIDs))
		seen := map[int64]bool{}
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
			if !seen[unitID] {
				seen[unitID] = true
				units = append(units, unit)
			}
		}

		rec := ballotRecord{
			GatheringID:               int64(gatheringID),
			ParticipantType:           ballotReq.VoterType,
			ParticipantName:           owner.Name,
			ParticipantIdentification: owner.IdentificationNumber,
			Units:                     units,
			Content:                   content,
			IP:                        req.RemoteAddr,
			UserAgent:                 req.UserAgent(),
		}
		if ballotReq.VoterType == "owner" {
			rec.OwnerID = sql.NullInt64{Int64: effectiveOwnerID, Valid: true}
		} else {
			rec.DelegatingOwnerID = sql.NullInt64{Int64: effectiveOwnerID, Valid: true}
			rec.DelegationDocumentRef = sql.NullString{String: strings.TrimSpace(ballotReq.DelegationDocumentRef), Valid: true}
			rec.ParticipantIdentification = rec.DelegationDocumentRef.String
			if name := strings.TrimSpace(ballotReq.DelegateName); name != "" {
				rec.ParticipantName = name
				rec.ParticipantIdentification = strings.TrimSpace(ballotReq.DelegateIdentification)
			}
		}

		ballot, participant, err := recordBallot(req.Context(), h.cfg, rec)
		if errors.Is(err, errUnitTaken) {
			handlers.RespondWithError(rw, http.StatusConflict, err.Error())
			return
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error recording ballot", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to submit ballot")
			return
		}
		afterBallot(req.Context(), h.cfg, int64(gatheringID), int64(associationID), participant.ID)

		// Log audit
		h.cfg.Db.CreateAuditLog(req.Context(), database.CreateAuditLogParams{
			GatheringID: int64(gatheringID),
			EntityType:  "ballot",
			EntityID:    ballot.ID,
			Action:      "submitted",
			PerformedBy: sql.NullString{String: handlers.GetUserIdFromContext(req), Valid: true},
			IpAddress:   sql.NullString{String: req.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"hash":"%s","voter_type":"%s","participant_id":%d}`, ballot.BallotHash, ballotReq.VoterType, participant.ID), Valid: true},
		})

		handlers.RespondWithJSON(rw, http.StatusCreated, map[string]interface{}{
			"status":         "ballot_submitted",
			"ballot_hash":    ballot.BallotHash,
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
			ID              int64  `json:"id"`
			ParticipantName string `json:"participant_name"`
			ParticipantType string `json:"participant_type"`
			// For a delegate: the owner they voted for.
			DelegatingOwnerName string  `json:"delegating_owner_name,omitempty"`
			UnitsInfo           []int64 `json:"units_info"`
			UnitsArea           float64 `json:"units_area"`
			UnitsPart           float64 `json:"units_part"`
			BallotHash          string  `json:"ballot_hash"`
			SubmittedAt         string  `json:"submitted_at"`
			IsValid             bool    `json:"is_valid"`
		}

		response := make([]BallotMetadata, len(ballots))
		for i, b := range ballots {
			submittedAt := ""
			if b.SubmittedAt.Valid {
				submittedAt = b.SubmittedAt.Time.Format(time.RFC3339)
			}

			var unitIDs []int64
			json.Unmarshal([]byte(b.UnitsInfo), &unitIDs)

			response[i] = BallotMetadata{
				ID:                  b.ID,
				ParticipantName:     b.ParticipantName,
				ParticipantType:     b.ParticipantType,
				DelegatingOwnerName: b.DelegatingOwnerName,
				UnitsInfo:           unitIDs,
				UnitsArea:           b.UnitsArea,
				UnitsPart:           b.UnitsPart,
				BallotHash:          b.BallotHash,
				SubmittedAt:         submittedAt,
				IsValid:             b.IsValid.Bool,
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

		// By the ballot's id (this used to pass the ballot id as a gathering
		// id, so it only ever found ballots whose ids matched one).
		ballot, err := h.cfg.Db.GetBallotByID(req.Context(), verifyReq.BallotID)
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Ballot not found")
			return
		}

		// Verify hash
		hash := sha256.Sum256([]byte(ballot.BallotContent))
		calculatedHash := hex.EncodeToString(hash[:])

		handlers.RespondWithJSON(rw, http.StatusOK, map[string]interface{}{
			"valid":        calculatedHash == verifyReq.BallotHash && ballot.BallotHash == verifyReq.BallotHash,
			"ballot_id":    ballot.ID,
			"submitted_at": domain.NullTimeToPtr(ballot.SubmittedAt),
			"is_valid":     ballot.IsValid.Bool,
		})
	}
}
