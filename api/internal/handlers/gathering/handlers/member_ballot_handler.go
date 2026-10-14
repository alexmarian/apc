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
// Accepts { ballot_content: { [matter_id]: BallotVote } } for the link's owner
// (see submitOwnerBallot).
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
		respondOwnerBallot(w, r, h.cfg, gathering, inv.OwnerID, fmt.Sprintf("member_owner_%d", inv.OwnerID), "member_app")
	}
}

// respondOwnerBallot decodes { ballot_content } and records it for owner
// (see submitOwnerBallot): 201 with the receipt, or the error.
func respondOwnerBallot(w http.ResponseWriter, r *http.Request, cfg *handlers.ApiConfig, gathering database.Gathering, ownerID int64, performedBy, source string) {
	var req struct {
		BallotContent map[string]domain.BallotVote `json:"ballot_content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handlers.RespondWithError(w, http.StatusBadRequest, "invalid request format")
		return
	}
	status, body := submitOwnerBallot(r, cfg, gathering, ownerID, req.BallotContent, performedBy, source)
	if msg, isErr := body.(string); isErr {
		handlers.RespondWithError(w, status, msg)
		return
	}
	handlers.RespondWithJSON(w, status, body)
}

// submitOwnerBallot records an owner's own ballot — from a member link or a
// signed-in owner account, the same way: the gathering must be active, the
// owner mustn't have voted, all their still-available units vote, the
// content must answer the agenda, and the participant is checked in. It
// returns the HTTP status and either the receipt or an error message.
func submitOwnerBallot(r *http.Request, cfg *handlers.ApiConfig, gathering database.Gathering, ownerID int64, ballotContent map[string]domain.BallotVote, performedBy, source string) (int, any) {
	ctx := r.Context()
	if gathering.Status != "active" {
		return http.StatusBadRequest, "gathering is not active"
	}
	// 409 if a ballot already exists for this owner in this gathering
	if existing, err := cfg.Db.GetGatheringParticipantByOwner(ctx, database.GetGatheringParticipantByOwnerParams{
		GatheringID: gathering.ID,
		OwnerID:     sql.NullInt64{Int64: ownerID, Valid: true},
	}); err == nil {
		if _, err := cfg.Db.GetBallotByParticipant(ctx, database.GetBallotByParticipantParams{GatheringID: gathering.ID, ParticipantID: existing.ID}); err == nil {
			return http.StatusConflict, "ballot already submitted"
		}
	}
	owner, err := cfg.Db.GetOwnerById(ctx, ownerID)
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to get owner", zap.Error(err))
		return http.StatusInternalServerError, "failed to load owner"
	}
	eligibleRows, err := cfg.Db.GetEligibleVotersWithUnits(ctx, database.GetEligibleVotersWithUnitsParams{
		GatheringID:   gathering.ID,
		AssociationID: gathering.AssociationID,
	})
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to get eligible units", zap.Error(err))
		return http.StatusInternalServerError, "failed to load units"
	}
	var units []database.GetEligibleVotersWithUnitsRow
	for _, row := range eligibleRows {
		if row.OwnerID == ownerID && row.IsAvailable == 1 {
			units = append(units, row)
		}
	}
	if len(units) == 0 {
		return http.StatusBadRequest, "no available units for voting"
	}
	matters, err := cfg.Db.GetVotingMatters(ctx, gathering.ID)
	if err != nil {
		return http.StatusInternalServerError, "failed to load voting matters"
	}
	content, err := validateBallotContent(matters, ballotContent)
	if err != nil {
		return http.StatusBadRequest, err.Error()
	}

	// Participant, auto check-in (no prior check-in step), unit slots and
	// ballot: all or nothing (see recordBallot).
	ballot, participant, err := recordBallot(ctx, cfg, ballotRecord{
		GatheringID:               gathering.ID,
		ParticipantType:           "owner",
		ParticipantName:           owner.Name,
		ParticipantIdentification: owner.IdentificationNumber,
		OwnerID:                   sql.NullInt64{Int64: ownerID, Valid: true},
		Units:                     units,
		Content:                   content,
		CheckIn:                   true,
		IP:                        r.RemoteAddr,
		UserAgent:                 r.UserAgent(),
	})
	if errors.Is(err, errUnitTaken) {
		return http.StatusConflict, err.Error()
	}
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "failed to record ballot", zap.Error(err))
		return http.StatusInternalServerError, "failed to submit ballot"
	}
	afterBallot(ctx, cfg, gathering.ID, gathering.AssociationID, participant.ID)

	cfg.Db.CreateAuditLog(ctx, database.CreateAuditLogParams{
		GatheringID: gathering.ID,
		EntityType:  "ballot",
		EntityID:    ballot.ID,
		Action:      "submitted",
		PerformedBy: sql.NullString{String: performedBy, Valid: true},
		IpAddress:   sql.NullString{String: r.RemoteAddr, Valid: true},
		Details:     sql.NullString{String: fmt.Sprintf(`{"hash":"%s","voter_type":"owner","source":"%s"}`, ballot.BallotHash, source), Valid: true},
	})

	var submittedAt *time.Time
	if ballot.SubmittedAt.Valid {
		submittedAt = &ballot.SubmittedAt.Time
	}
	return http.StatusCreated, map[string]interface{}{
		"ballot_id":    ballot.ID,
		"ballot_hash":  ballot.BallotHash,
		"submitted_at": submittedAt,
	}
}
