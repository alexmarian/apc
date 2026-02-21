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

	// Get all voting matters for this gathering
	matters, err := s.db.GetVotingMatters(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get voting matters for tally update",
			zap.Int64("gathering_id", gatheringID),
			zap.Error(err))
		return
	}

	// Get all ballots
	ballots, err := s.db.GetBallotsForGathering(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get ballots for tally update",
			zap.Int64("gathering_id", gatheringID),
			zap.Error(err))
		return
	}

	// Get all participants to get voting weights
	participants, err := s.db.GetGatheringParticipants(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.ErrorLevel, "Failed to get participants for tally update",
			zap.Int64("gathering_id", gatheringID),
			zap.Error(err))
		return
	}

	// Create participant weight lookup map
	participantWeights := make(map[int64]float64)
	for _, p := range participants {
		participantWeights[p.ID] = p.UnitsPart
	}

	// Process each voting matter
	for _, matter := range matters {
		var votingConfig domain.VotingConfig
		if err := json.Unmarshal([]byte(matter.VotingConfig), &votingConfig); err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to unmarshal voting config",
				zap.Int64("matter_id", matter.ID),
				zap.Error(err))
			continue
		}

		// Initialize tally
		tally := make(map[string]domain.TallyResult)

		// Initialize based on voting type
		if votingConfig.Type == "yes_no" {
			tally["yes"] = domain.TallyResult{Count: 0, Weight: 0}
			tally["no"] = domain.TallyResult{Count: 0, Weight: 0}
			if votingConfig.AllowAbstention {
				tally["abstain"] = domain.TallyResult{Count: 0, Weight: 0}
			}
		} else if votingConfig.Type == "single_choice" || votingConfig.Type == "multiple_choice" {
			for _, option := range votingConfig.Options {
				tally[option.ID] = domain.TallyResult{Count: 0, Weight: 0}
			}
			if votingConfig.AllowAbstention {
				tally["abstain"] = domain.TallyResult{Count: 0, Weight: 0}
			}
		}

		// Count votes from valid ballots
		for _, ballot := range ballots {
			if !ballot.IsValid.Bool {
				continue
			}

			participantWeight := participantWeights[ballot.ParticipantID]

			// Parse ballot content
			var ballotContent map[string]domain.BallotVote
			if err := json.Unmarshal([]byte(ballot.BallotContent), &ballotContent); err != nil {
				logging.Logger.Log(zap.WarnLevel, "Failed to unmarshal ballot content",
					zap.Int64("ballot_id", ballot.ID),
					zap.Error(err))
				continue
			}

			// Find vote for this matter
			matterIDStr := strconv.FormatInt(matter.ID, 10)
			if vote, ok := ballotContent[matterIDStr]; ok {
				voteKey := vote.VoteValue
				if vote.OptionID != "" {
					voteKey = vote.OptionID
				}

				if currentTally, exists := tally[voteKey]; exists {
					tally[voteKey] = domain.TallyResult{
						Count:  currentTally.Count + 1,
						Weight: currentTally.Weight + participantWeight,
					}
				}
			}
		}

		// Calculate percentages for each option
		totalCount := 0
		totalWeight := 0.0
		for _, t := range tally {
			totalCount += t.Count
			totalWeight += t.Weight
		}
		for key, t := range tally {
			countPct := 0.0
			weightPct := 0.0
			if totalCount > 0 {
				countPct = float64(t.Count) / float64(totalCount) * 100
			}
			if totalWeight > 0 {
				weightPct = t.Weight / totalWeight * 100
			}
			tally[key] = domain.TallyResult{
				Count:           t.Count,
				Weight:          t.Weight,
				Area:            t.Area,
				Percentage:      countPct,
				WeightPercentage: weightPct,
			}
		}

		// Store tally in database
		tallyJSON, err := json.Marshal(tally)
		if err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to marshal tally data",
				zap.Int64("matter_id", matter.ID),
				zap.Error(err))
			continue
		}

		// Upsert tally
		_, err = s.db.UpsertVoteTally(ctx, database.UpsertVoteTallyParams{
			GatheringID:    gatheringID,
			VotingMatterID: matter.ID,
			TallyData:      string(tallyJSON),
		})
		if err != nil {
			logging.Logger.Log(zap.ErrorLevel, "Failed to upsert vote tally",
				zap.Int64("matter_id", matter.ID),
				zap.Error(err))
		}
	}

	logging.Logger.Log(zap.InfoLevel, "Vote tallies updated",
		zap.Int64("gathering_id", gatheringID),
		zap.Int("matter_count", len(matters)))
}
