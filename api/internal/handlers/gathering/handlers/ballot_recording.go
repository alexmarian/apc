package handlers

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/services"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// errUnitTaken: a unit's voting slot was claimed by someone else (e.g. the
// owner voted from the member app meanwhile). Nothing was recorded.
var errUnitTaken = errors.New("voting rights for this unit have already been exercised")

// ballotRecord is one ballot to record, with the participant casting it.
type ballotRecord struct {
	GatheringID               int64
	ParticipantType           string // owner | delegate
	ParticipantName           string
	ParticipantIdentification string
	OwnerID                   sql.NullInt64
	DelegatingOwnerID         sql.NullInt64
	DelegationDocumentRef     sql.NullString
	Units                     []database.GetEligibleVotersWithUnitsRow
	Content                   map[string]domain.BallotVote
	CheckIn                   bool
	IP, UserAgent             string
}

// recordBallot creates the participant, claims every unit's voting slot and
// stores the ballot in one transaction: if any slot is already taken nothing
// is kept (errUnitTaken), so a unit can't be voted twice and a failed
// attempt doesn't leave units claimed.
func recordBallot(ctx context.Context, cfg *handlers.ApiConfig, rec ballotRecord) (database.VotingBallot, database.GatheringParticipant, error) {
	var ballot database.VotingBallot
	var participant database.GatheringParticipant

	unitIDs := make([]int64, len(rec.Units))
	area, weight := 0.0, 0.0
	for i, u := range rec.Units {
		unitIDs[i] = u.UnitID
		area += u.Area
		weight += u.VotingWeight
	}
	unitsJSON, _ := json.Marshal(unitIDs)
	ballotJSON, err := json.Marshal(rec.Content) // map keys are sorted: a stable hash
	if err != nil {
		return ballot, participant, err
	}
	hash := sha256.Sum256(ballotJSON)

	tx, err := cfg.DB.BeginTx(ctx, nil)
	if err != nil {
		return ballot, participant, err
	}
	defer tx.Rollback()
	q := cfg.Db.WithTx(tx)

	participant, err = q.CreateGatheringParticipant(ctx, database.CreateGatheringParticipantParams{
		GatheringID:               rec.GatheringID,
		ParticipantType:           rec.ParticipantType,
		ParticipantName:           rec.ParticipantName,
		ParticipantIdentification: sql.NullString{String: rec.ParticipantIdentification, Valid: rec.ParticipantIdentification != ""},
		OwnerID:                   rec.OwnerID,
		DelegatingOwnerID:         rec.DelegatingOwnerID,
		DelegationDocumentRef:     rec.DelegationDocumentRef,
		UnitsInfo:                 string(unitsJSON),
		UnitsArea:                 area,
		UnitsPart:                 weight,
	})
	if err != nil {
		return ballot, participant, err
	}
	if rec.CheckIn {
		if _, err := q.CheckInParticipant(ctx, database.CheckInParticipantParams{ID: participant.ID, GatheringID: rec.GatheringID}); err != nil {
			return ballot, participant, err
		}
	}
	for _, unitID := range unitIDs {
		_, err := q.AssignUnitSlot(ctx, database.AssignUnitSlotParams{
			ParticipantID: participant.ID,
			GatheringID:   rec.GatheringID,
			UnitID:        unitID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			return ballot, participant, errUnitTaken
		}
		if err != nil {
			return ballot, participant, err
		}
	}
	ballot, err = q.CreateBallot(ctx, database.CreateBallotParams{
		GatheringID:        rec.GatheringID,
		ParticipantID:      participant.ID,
		BallotContent:      string(ballotJSON),
		BallotHash:         hex.EncodeToString(hash[:]),
		SubmittedIp:        sql.NullString{String: rec.IP, Valid: true},
		SubmittedUserAgent: sql.NullString{String: rec.UserAgent, Valid: true},
	})
	if err != nil {
		return ballot, participant, err
	}
	return ballot, participant, tx.Commit()
}

// afterBallot refreshes what a new ballot changes: participation stats and
// tallies first, then drops the cached results — in that order, so a results
// request can't recompute (and cache) from the tallies before this ballot.
func afterBallot(ctx context.Context, cfg *handlers.ApiConfig, gatheringID, associationID, participantID int64) {
	services.NewStatsService(cfg.Db).UpdateGatheringParticipationStats(gatheringID, associationID)
	services.NewTallyService(cfg.Db).UpdateVoteTallies(gatheringID, int(participantID))
	quorum, tally := services.NewQuorumService(cfg.Db), services.NewTallyService(cfg.Db)
	if err := services.NewVotingResultsService(cfg.Db, quorum, tally).InvalidateResults(ctx, gatheringID); err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error invalidating cached results", zap.Error(err))
	}
}

// validateBallotContent checks a ballot against the gathering's matters and
// returns it normalized (keyed by matter id, informative matters dropped):
// every voting matter needs a valid answer — yes/no or one option (single
// choice), distinct options (multiple choice), every option once in order
// of preference (ranking) — and "abstain" alone where abstention is allowed.
func validateBallotContent(matters []database.VotingMatter, content map[string]domain.BallotVote) (map[string]domain.BallotVote, error) {
	known := map[string]bool{}
	normalized := map[string]domain.BallotVote{}
	for _, dbMatter := range matters {
		m := domain.DBVotingMatterToResponse(dbMatter)
		key := strconv.FormatInt(m.ID, 10)
		known[key] = true
		if m.IsInformative {
			continue
		}
		values := content[key].Values
		if len(values) == 0 {
			return nil, fmt.Errorf("Missing vote for %q", m.Title)
		}
		invalid := fmt.Errorf("Invalid vote for %q", m.Title)
		if len(values) == 1 && values[0] == "abstain" {
			if !m.VotingConfig.AllowAbstention {
				return nil, fmt.Errorf("Abstention is not allowed for %q", m.Title)
			}
			normalized[key] = domain.BallotVote{MatterID: m.ID, Values: values}
			continue
		}
		options := make([]string, len(m.VotingConfig.Options))
		for i, o := range m.VotingConfig.Options {
			options[i] = o.ID
		}
		distinct := map[string]bool{}
		for _, v := range values {
			if distinct[v] {
				return nil, invalid
			}
			distinct[v] = true
		}
		switch m.VotingConfig.Type {
		case "yes_no":
			if len(values) != 1 || (values[0] != "yes" && values[0] != "no") {
				return nil, invalid
			}
		case "single_choice":
			if len(values) != 1 || !slices.Contains(options, values[0]) {
				return nil, invalid
			}
		case "multiple_choice":
			for _, v := range values {
				if !slices.Contains(options, v) {
					return nil, invalid
				}
			}
		case "ranking":
			if len(values) != len(options) {
				return nil, invalid
			}
			for _, v := range values {
				if !slices.Contains(options, v) {
					return nil, invalid
				}
			}
		default:
			return nil, invalid
		}
		normalized[key] = domain.BallotVote{MatterID: m.ID, Values: values}
	}
	for key := range content {
		if !known[key] {
			return nil, fmt.Errorf("Unknown voting matter %s", key)
		}
	}
	return normalized, nil
}
