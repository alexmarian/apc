package services

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// TallyService handles vote tallying and counting
type TallyService struct {
	db *database.Queries
}

// NewTallyService creates a new TallyService
func NewTallyService(db *database.Queries) *TallyService {
	return &TallyService{db: db}
}

// UpdateVoteTallies updates the vote tallies for all matters in a gathering
func (s *TallyService) UpdateVoteTallies(gatheringID int64, participantID int) {
	ctx := context.Background()

	matters, err := s.db.GetVotingMatters(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get voting matters for tally update",
			zap.Int64("gathering_id", gatheringID), zap.Error(err))
		return
	}

	ballots, err := s.db.GetBallotsForGathering(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get ballots for tally update",
			zap.Int64("gathering_id", gatheringID), zap.Error(err))
		return
	}

	participants, err := s.db.GetGatheringParticipants(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get participants for tally update",
			zap.Int64("gathering_id", gatheringID), zap.Error(err))
		return
	}

	participantWeights := make(map[int64]float64)
	for _, p := range participants {
		participantWeights[p.ID] = p.UnitsPart
	}

	for _, matter := range matters {
		var votingConfig domain.VotingConfig
		if err := json.Unmarshal([]byte(matter.VotingConfig), &votingConfig); err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to unmarshal voting config",
				zap.Int64("matter_id", matter.ID), zap.Error(err))
			continue
		}

		tally := initTally(votingConfig)

		for _, ballot := range ballots {
			if !ballot.IsValid.Bool {
				continue
			}

			weight := participantWeights[ballot.ParticipantID]

			var ballotContent map[string]domain.BallotVote
			if err := json.Unmarshal([]byte(ballot.BallotContent), &ballotContent); err != nil {
				logging.Logger.Log(zap.WarnLevel, "Failed to unmarshal ballot content",
					zap.Int64("ballot_id", ballot.ID), zap.Error(err))
				continue
			}

			matterIDStr := strconv.FormatInt(matter.ID, 10)
			vote, ok := ballotContent[matterIDStr]
			if !ok || len(vote.Values) == 0 {
				continue
			}

			switch votingConfig.Type {
			case "yes_no", "single_choice":
				key := vote.Values[0]
				if t, exists := tally[key]; exists {
					tally[key] = domain.TallyResult{Count: t.Count + 1, Weight: t.Weight + weight}
				}

			case "multiple_choice":
				for _, optID := range vote.Values {
					if t, exists := tally[optID]; exists {
						tally[optID] = domain.TallyResult{Count: t.Count + 1, Weight: t.Weight + weight}
					}
				}

			case "ranking":
				// Borda count: first choice gets N-1 points, last gets 0.
				// IRV and first-choice plurality are alternative methods not used here.
				n := len(vote.Values)
				for i, optID := range vote.Values {
					points := n - 1 - i
					if t, exists := tally[optID]; exists {
						tally[optID] = domain.TallyResult{
							Count:  t.Count + points,
							Weight: t.Weight + weight*float64(points),
						}
					}
				}
			}
		}

		// Calculate percentages
		totalCount := 0
		totalWeight := 0.0
		for _, t := range tally {
			totalCount += t.Count
			totalWeight += t.Weight
		}
		for key, t := range tally {
			countPct, weightPct := 0.0, 0.0
			if totalCount > 0 {
				countPct = float64(t.Count) / float64(totalCount) * 100
			}
			if totalWeight > 0 {
				weightPct = t.Weight / totalWeight * 100
			}
			tally[key] = domain.TallyResult{
				Count:            t.Count,
				Weight:           t.Weight,
				Area:             t.Area,
				Percentage:       countPct,
				WeightPercentage: weightPct,
			}
		}

		tallyJSON, err := json.Marshal(tally)
		if err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to marshal tally data",
				zap.Int64("matter_id", matter.ID), zap.Error(err))
			continue
		}

		_, err = s.db.UpsertVoteTally(ctx, database.UpsertVoteTallyParams{
			GatheringID:    gatheringID,
			VotingMatterID: matter.ID,
			TallyData:      string(tallyJSON),
		})
		if err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to upsert vote tally",
				zap.Int64("matter_id", matter.ID), zap.Error(err))
		}
	}

	logging.Logger.Log(zap.InfoLevel, "Vote tallies updated",
		zap.Int64("gathering_id", gatheringID), zap.Int("matter_count", len(matters)))
}

func initTally(config domain.VotingConfig) map[string]domain.TallyResult {
	tally := make(map[string]domain.TallyResult)
	switch config.Type {
	case "yes_no":
		tally["yes"] = domain.TallyResult{}
		tally["no"] = domain.TallyResult{}
		if config.AllowAbstention {
			tally["abstain"] = domain.TallyResult{}
		}
	case "single_choice", "multiple_choice", "ranking":
		for _, opt := range config.Options {
			tally[opt.ID] = domain.TallyResult{}
		}
		if config.AllowAbstention {
			tally["abstain"] = domain.TallyResult{}
		}
	}
	return tally
}
