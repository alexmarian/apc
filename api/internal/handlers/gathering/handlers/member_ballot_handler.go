package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/services"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// MemberBallotHandler handles ballot submission for the member app.
type MemberBallotHandler struct {
	cfg                  *handlers.ApiConfig
	statsService         *services.StatsService
	tallyService         *services.TallyService
	votingResultsService *services.VotingResultsService
}

// NewMemberBallotHandler creates a new MemberBallotHandler.
func NewMemberBallotHandler(cfg *handlers.ApiConfig) *MemberBallotHandler {
	tallyService := services.NewTallyService(cfg.Db)
	quorumService := services.NewQuorumService(cfg.Db)
	return &MemberBallotHandler{
		cfg:                  cfg,
		statsService:         services.NewStatsService(cfg.Db),
		tallyService:         tallyService,
		votingResultsService: services.NewVotingResultsService(cfg.Db, quorumService, tallyService),
	}
}

// HandleSubmitMemberBallot handles POST /v1/api/member/gatherings/{memberToken}/ballot.
// Accepts { ballot_content: { [matter_id]: BallotVote } }, creates a participant
// (with auto check-in), assigns unit slots, stores the ballot, and triggers async tallying.
func (h *MemberBallotHandler) HandleSubmitMemberBallot() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inv, ok := handlers.MemberInvitationFromContext(r.Context())
		if !ok {
			handlers.RespondWithError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if inv.RevokedAt != nil {
			handlers.RespondWithError(w, http.StatusUnauthorized, "invitation has been revoked")
			return
		}
		if time.Now().After(inv.ExpiresAt) {
			handlers.RespondWithError(w, http.StatusUnauthorized, "invitation has expired")
			return
		}

		gathering, err := h.cfg.Db.GetGatheringByID(r.Context(), inv.GatheringID)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to get gathering", zap.Error(err))
			handlers.RespondWithError(w, http.StatusInternalServerError, "failed to load gathering")
			return
		}
		if gathering.Status != "active" {
			handlers.RespondWithError(w, http.StatusBadRequest, "gathering is not active")
			return
		}

		// 409 if a ballot already exists for this owner in this gathering
		existingParticipant, err := h.cfg.Db.GetGatheringParticipantByOwner(r.Context(), database.GetGatheringParticipantByOwnerParams{
			GatheringID: inv.GatheringID,
			OwnerID:     sql.NullInt64{Int64: inv.OwnerID, Valid: true},
		})
		if err == nil {
			_, err := h.cfg.Db.GetBallotByParticipant(r.Context(), database.GetBallotByParticipantParams{
				GatheringID:   inv.GatheringID,
				ParticipantID: existingParticipant.ID,
			})
			if err == nil {
				handlers.RespondWithError(w, http.StatusConflict, "ballot already submitted")
				return
			}
		}

		var req struct {
			BallotContent map[string]domain.BallotVote `json:"ballot_content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			handlers.RespondWithError(w, http.StatusBadRequest, "invalid request format")
			return
		}

		owner, err := h.cfg.Db.GetOwnerById(r.Context(), inv.OwnerID)
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to get owner", zap.Error(err))
			handlers.RespondWithError(w, http.StatusInternalServerError, "failed to load owner")
			return
		}

		eligibleRows, err := h.cfg.Db.GetEligibleVotersWithUnits(r.Context(), database.GetEligibleVotersWithUnitsParams{
			GatheringID:   inv.GatheringID,
			AssociationID: gathering.AssociationID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to get eligible units", zap.Error(err))
			handlers.RespondWithError(w, http.StatusInternalServerError, "failed to load units")
			return
		}

		var units []database.GetEligibleVotersWithUnitsRow
		for _, row := range eligibleRows {
			if row.OwnerID == inv.OwnerID && row.IsAvailable == 1 {
				units = append(units, row)
			}
		}
		if len(units) == 0 {
			handlers.RespondWithError(w, http.StatusBadRequest, "no available units for voting")
			return
		}

		matters, err := h.cfg.Db.GetVotingMatters(r.Context(), inv.GatheringID)
		if err != nil {
			handlers.RespondWithError(w, http.StatusInternalServerError, "failed to load voting matters")
			return
		}
		content, err := validateBallotContent(matters, req.BallotContent)
		if err != nil {
			handlers.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Participant, auto check-in (no prior check-in step), unit slots and
		// ballot: all or nothing (see recordBallot).
		ballot, participant, err := recordBallot(r.Context(), h.cfg, ballotRecord{
			GatheringID:               inv.GatheringID,
			ParticipantType:           "owner",
			ParticipantName:           owner.Name,
			ParticipantIdentification: owner.IdentificationNumber,
			OwnerID:                   sql.NullInt64{Int64: inv.OwnerID, Valid: true},
			Units:                     units,
			Content:                   content,
			CheckIn:                   true,
			IP:                        r.RemoteAddr,
			UserAgent:                 r.UserAgent(),
		})
		if errors.Is(err, errUnitTaken) {
			handlers.RespondWithError(w, http.StatusConflict, err.Error())
			return
		}
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "failed to record ballot", zap.Error(err))
			handlers.RespondWithError(w, http.StatusInternalServerError, "failed to submit ballot")
			return
		}
		ballotHash := ballot.BallotHash
		afterBallot(r.Context(), h.cfg, inv.GatheringID, gathering.AssociationID, participant.ID)

		h.cfg.Db.CreateAuditLog(r.Context(), database.CreateAuditLogParams{
			GatheringID: inv.GatheringID,
			EntityType:  "ballot",
			EntityID:    ballot.ID,
			Action:      "submitted",
			PerformedBy: sql.NullString{String: fmt.Sprintf("member_owner_%d", inv.OwnerID), Valid: true},
			IpAddress:   sql.NullString{String: r.RemoteAddr, Valid: true},
			Details:     sql.NullString{String: fmt.Sprintf(`{"hash":"%s","voter_type":"owner","source":"member_app"}`, ballotHash), Valid: true},
		})

		var submittedAt *time.Time
		if ballot.SubmittedAt.Valid {
			submittedAt = &ballot.SubmittedAt.Time
		}

		handlers.RespondWithJSON(w, http.StatusCreated, map[string]interface{}{
			"ballot_id":    ballot.ID,
			"ballot_hash":  ballotHash,
			"submitted_at": submittedAt,
		})
	}
}
