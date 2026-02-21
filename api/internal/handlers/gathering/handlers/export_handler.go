package handlers

import (
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

// ExportHandler handles export operations (markdown reports, etc.)
type ExportHandler struct {
	cfg           *handlers.ApiConfig
	quorumService *services.QuorumService
}

// NewExportHandler creates a new ExportHandler
func NewExportHandler(cfg *handlers.ApiConfig) *ExportHandler {
	return &ExportHandler{
		cfg:           cfg,
		quorumService: services.NewQuorumService(cfg.Db),
	}
}

// HandleDownloadVotingResults generates and downloads a markdown report of voting results
func (h *ExportHandler) HandleDownloadVotingResults() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get gathering details
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Get all voting matters
		matters, err := h.cfg.Db.GetVotingMatters(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting voting matters", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting matters")
			return
		}

		// Get all ballots
		ballots, err := h.cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballots", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get ballots")
			return
		}

		// Get participants
		participants, _ := h.cfg.Db.GetGatheringParticipants(req.Context(), int64(gatheringID))

		// Get voted units stats
		votedStats, _ := h.cfg.Db.GetVotedUnitsStats(req.Context(), int64(gatheringID))
		votedUnitsPart, _ := votedStats.VotedUnitsTotalPart.(float64)
		votedUnitsArea, _ := votedStats.VotedUnitsTotalArea.(float64)

		// Build markdown report
		var md string
		md += fmt.Sprintf("# Voting Results: %s\n\n", gathering.Title)
		md += fmt.Sprintf("**Date:** %s\n\n", gathering.GatheringDate.Format("2006-01-02 15:04"))
		md += fmt.Sprintf("**Location:** %s\n\n", gathering.Location)
		md += fmt.Sprintf("**Status:** %s\n\n", gathering.Status)
		if gathering.Status == "closed" || gathering.Status == "tallied" {
			md += fmt.Sprintf("**Closed At:** %s\n\n", gathering.UpdatedAt.Time.Format("2006-01-02 15:04"))
		}

		md += "## Participation Statistics\n\n"
		md += "| Metric | Count | Weight | Area (m²) |\n"
		md += "|--------|-------|--------|----------|\n"
		md += fmt.Sprintf("| **Qualified Units** | %d | %.4f | %.2f |\n",
			gathering.QualifiedUnitsCount.Int64,
			gathering.QualifiedUnitsTotalPart.Float64,
			gathering.QualifiedUnitsTotalArea.Float64)
		md += fmt.Sprintf("| **Participating Units** | %d | %.4f | %.2f |\n",
			gathering.ParticipatingUnitsCount.Int64,
			gathering.ParticipatingUnitsTotalPart.Float64,
			gathering.ParticipatingUnitsTotalArea.Float64)
		md += fmt.Sprintf("| **Voted Units** | %d | %.4f | %.2f |\n\n",
			votedStats.VotedUnitsCount,
			votedUnitsPart,
			votedUnitsArea)

		// Calculate rates
		participationRate := 0.0
		votingRate := 0.0
		if gathering.QualifiedUnitsTotalPart.Float64 > 0 {
			participationRate = (gathering.ParticipatingUnitsTotalPart.Float64 / gathering.QualifiedUnitsTotalPart.Float64) * 100
			votingRate = (votedUnitsPart / gathering.QualifiedUnitsTotalPart.Float64) * 100
		}

		md += fmt.Sprintf("**Participation Rate:** %.2f%% (by weight)\n\n", participationRate)
		md += fmt.Sprintf("**Voting Completion Rate:** %.2f%% (by weight)\n\n", votingRate)

		md += "## Voting Matters and Results\n\n"

		// Process each voting matter
		for _, matter := range matters {
			var votingConfig domain.VotingConfig
			json.Unmarshal([]byte(matter.VotingConfig), &votingConfig)

			md += fmt.Sprintf("### %d. %s\n\n", matter.OrderIndex, matter.Title)
			if matter.Description.Valid && matter.Description.String != "" {
				md += fmt.Sprintf("**Description:** %s\n\n", matter.Description.String)
			}
			md += fmt.Sprintf("**Type:** %s\n\n", matter.MatterType)
			md += fmt.Sprintf("**Voting Method:** %s\n\n", votingConfig.Type)
			md += fmt.Sprintf("**Required Majority:** %s\n\n", votingConfig.RequiredMajority)

			// Calculate tally
			tally := make(map[string]domain.TallyResult)

			// Initialize options based on voting type
			if votingConfig.Type == "yes_no" {
				tally["yes"] = domain.TallyResult{Count: 0, Weight: 0, Area: 0}
				tally["no"] = domain.TallyResult{Count: 0, Weight: 0, Area: 0}
				if votingConfig.AllowAbstention {
					tally["abstain"] = domain.TallyResult{Count: 0, Weight: 0, Area: 0}
				}
			} else if votingConfig.Type == "multiple_choice" || votingConfig.Type == "single_choice" {
				for _, option := range votingConfig.Options {
					tally[option.ID] = domain.TallyResult{Count: 0, Weight: 0, Area: 0}
				}
				if votingConfig.AllowAbstention {
					tally["abstain"] = domain.TallyResult{Count: 0, Weight: 0, Area: 0}
				}
			}

			// Count votes from valid ballots
			totalWeight := 0.0
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
				var ballotContent map[string]domain.BallotVote
				json.Unmarshal([]byte(ballot.BallotContent), &ballotContent)

				// Find vote for this matter
				matterIDStr := strconv.FormatInt(matter.ID, 10)
				if vote, ok := ballotContent[matterIDStr]; ok {
					voteKey := vote.VoteValue
					if vote.OptionID != "" {
						voteKey = vote.OptionID
					}

					if _, exists := tally[voteKey]; exists {
						tally[voteKey] = domain.TallyResult{
							Count:  tally[voteKey].Count + 1,
							Weight: tally[voteKey].Weight + participantWeight,
							Area:   tally[voteKey].Area + participantArea,
						}
						totalWeight += participantWeight
					}
				}
			}

			// Display results
			md += "**Results:**\n\n"
			md += "| Option | Votes | % Votes | Weight | % Weight (of cast) | % Weight (of qualified) |\n"
			md += "|--------|-------|---------|--------|--------------------|------------------------|\n"

			totalTallyCount := 0
			totalTallyWeight := 0.0
			for _, r := range tally {
				totalTallyCount += r.Count
				totalTallyWeight += r.Weight
			}

			for key, result := range tally {
				countPct := 0.0
				weightPctOfCast := 0.0
				weightPctOfQualified := 0.0
				if totalTallyCount > 0 {
					countPct = float64(result.Count) / float64(totalTallyCount) * 100
				}
				if totalTallyWeight > 0 {
					weightPctOfCast = result.Weight / totalTallyWeight * 100
				}
				if gathering.QualifiedUnitsTotalPart.Float64 > 0 {
					weightPctOfQualified = services.RoundTo3Decimals(result.Weight / gathering.QualifiedUnitsTotalPart.Float64 * 100)
				}

				displayKey := key
				if votingConfig.Type == "multiple_choice" || votingConfig.Type == "single_choice" {
					for _, opt := range votingConfig.Options {
						if opt.ID == key {
							displayKey = opt.Text
							break
						}
					}
				}

				md += fmt.Sprintf("| %s | %d | %.2f%% | %.4f | %.2f%% | %.3f%% |\n",
					displayKey, result.Count, countPct, result.Weight, weightPctOfCast, weightPctOfQualified)
			}
			md += "\n"

			// Determine if passed
			matterResult := domain.VoteMatterResult{
				MatterID:       matter.ID,
				MatterTitle:    matter.Title,
				MatterType:     matter.MatterType,
				VotingConfig:   votingConfig,
				Tally:          tally,
				TotalVoted:     totalWeight,
				TotalAbstained: tally["abstain"].Weight,
			}
			passed := h.quorumService.CalculateIfPassed(matterResult, votingConfig, gathering)

			if votingConfig.RequiredMajority == "informative" {
				md += "**Status:** Informative (no pass/fail)\n\n"
			} else if passed {
				md += "**Status:** ✅ PASSED\n\n"
			} else {
				md += "**Status:** ❌ FAILED\n\n"
			}

			md += "---\n\n"
		}

		md += fmt.Sprintf("*Report generated at: %s*\n", time.Now().Format("2006-01-02 15:04:05"))

		// Set headers for file download
		filename := fmt.Sprintf("voting-results-%s-%s.md",
			gathering.Title,
			time.Now().Format("2006-01-02"))
		rw.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(md))
	}
}

// HandleDownloadVotingBallots generates and downloads a markdown report of all ballots
func (h *ExportHandler) HandleDownloadVotingBallots() func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		associationID, _ := strconv.Atoi(req.PathValue(handlers.AssociationIdPathValue))
		gatheringID, _ := strconv.Atoi(req.PathValue(domain.GatheringIDPathValue))

		// Get gathering details
		gathering, err := h.cfg.Db.GetGathering(req.Context(), database.GetGatheringParams{
			ID:            int64(gatheringID),
			AssociationID: int64(associationID),
		})
		if err != nil {
			handlers.RespondWithError(rw, http.StatusNotFound, "Gathering not found")
			return
		}

		// Get all voting matters
		matters, err := h.cfg.Db.GetVotingMatters(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting voting matters", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get voting matters")
			return
		}

		// Create matter lookup map
		matterMap := make(map[int64]database.VotingMatter)
		for _, m := range matters {
			matterMap[m.ID] = m
		}

		// Get all ballots
		ballots, err := h.cfg.Db.GetBallotsForGathering(req.Context(), int64(gatheringID))
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error getting ballots", zap.Error(err))
			handlers.RespondWithError(rw, http.StatusInternalServerError, "Failed to get ballots")
			return
		}

		// Build markdown report
		var md string
		md += fmt.Sprintf("# Voting Ballots: %s\n\n", gathering.Title)
		md += fmt.Sprintf("**Date:** %s\n\n", gathering.GatheringDate.Format("2006-01-02 15:04"))
		md += fmt.Sprintf("**Total Ballots:** %d\n\n", len(ballots))

		md += "---\n\n"

		// List all ballots
		for i, ballot := range ballots {
			md += fmt.Sprintf("## Ballot #%d\n\n", i+1)
			md += fmt.Sprintf("**Participant:** %s\n\n", ballot.ParticipantName)
			md += fmt.Sprintf("**Units Weight:** %.4f\n\n", ballot.UnitsPart)
			md += fmt.Sprintf("**Units Area:** %.2f m²\n\n", ballot.UnitsArea)

			if ballot.SubmittedAt.Valid {
				md += fmt.Sprintf("**Submitted:** %s\n\n", ballot.SubmittedAt.Time.Format("2006-01-02 15:04:05"))
			}

			md += fmt.Sprintf("**Ballot Hash:** `%s`\n\n", ballot.BallotHash)
			md += fmt.Sprintf("**Valid:** %t\n\n", ballot.IsValid.Bool)

			if !ballot.IsValid.Bool {
				md += fmt.Sprintf("**Invalidation Reason:** %s\n\n", ballot.InvalidationReason.String)
			}

			// Parse ballot content
			var ballotContent map[string]domain.BallotVote
			if err := json.Unmarshal([]byte(ballot.BallotContent), &ballotContent); err == nil {
				md += "**Votes:**\n\n"

				for matterIDStr, vote := range ballotContent {
					matterID, _ := strconv.ParseInt(matterIDStr, 10, 64)
					matter, ok := matterMap[matterID]
					if !ok {
						continue
					}

					md += fmt.Sprintf("- **%s:** ", matter.Title)

					if vote.OptionID != "" {
						// Find option text
						var config domain.VotingConfig
						json.Unmarshal([]byte(matter.VotingConfig), &config)
						for _, opt := range config.Options {
							if opt.ID == vote.OptionID {
								md += opt.Text
								break
							}
						}
					} else {
						md += vote.VoteValue
					}
					md += "\n"
				}
				md += "\n"
			}

			md += "---\n\n"
		}

		md += fmt.Sprintf("*Report generated at: %s*\n", time.Now().Format("2006-01-02 15:04:05"))

		// Set headers for file download
		filename := fmt.Sprintf("voting-ballots-%s-%s.md",
			gathering.Title,
			time.Now().Format("2006-01-02"))
		rw.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		rw.WriteHeader(http.StatusOK)
		rw.Write([]byte(md))
	}
}
