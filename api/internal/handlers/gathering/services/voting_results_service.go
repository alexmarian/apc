package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// VotingResultsService handles computation and caching of voting results
type VotingResultsService struct {
	db            *database.Queries
	quorumService *QuorumService
	tallyService  *TallyService
}

// NewVotingResultsService creates a new VotingResultsService
func NewVotingResultsService(db *database.Queries, quorumService *QuorumService, tallyService *TallyService) *VotingResultsService {
	return &VotingResultsService{
		db:            db,
		quorumService: quorumService,
		tallyService:  tallyService,
	}
}

// ComputeAndStoreResults computes voting results for a gathering and caches them
func (s *VotingResultsService) ComputeAndStoreResults(ctx context.Context, gatheringID int64, associationID int64) (*domain.VoteResults, error) {
	log.Printf("[VotingResultsService] Computing results for gathering %d", gatheringID)

	// Get gathering details
	dbGathering, err := s.db.GetGathering(ctx, database.GetGatheringParams{
		ID:            gatheringID,
		AssociationID: associationID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get gathering: %w", err)
	}

	gathering := domain.DBGatheringToResponse(dbGathering)

	// Get voting strategy based on gathering voting mode
	strategy := GetVotingStrategy(gathering.VotingMode)

	// Get all voting matters for this gathering
	dbMatters, err := s.db.GetVotingMatters(ctx, gatheringID)
	if err != nil {
		return nil, fmt.Errorf("failed to get voting matters: %w", err)
	}

	// Get all tallies
	dbTallies, err := s.db.GetAllVoteTallies(ctx, gatheringID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tallies: %w", err)
	}

	// Get participation stats
	participatingStats, err := s.db.GetParticipatingUnitsStats(ctx, gatheringID)
	if err != nil {
		return nil, fmt.Errorf("failed to get participating stats: %w", err)
	}

	votedStats, err := s.db.GetVotedUnitsStats(ctx, gatheringID)
	if err != nil {
		return nil, fmt.Errorf("failed to get voted stats: %w", err)
	}

	// Extract and convert stats (SQLite returns interface{})
	participatingWeight := float64(0)
	if participatingStats.ParticipatingUnitsTotalPart != nil {
		participatingWeight = participatingStats.ParticipatingUnitsTotalPart.(float64)
	}

	votedWeight := float64(0)
	if votedStats.VotedUnitsTotalPart != nil {
		votedWeight = votedStats.VotedUnitsTotalPart.(float64)
	}

	// Calculate quorum
	quorumInfo := s.quorumService.CalculateQuorum(
		gathering,
		participatingWeight,
		int(participatingStats.ParticipatingUnitsCount),
		votedWeight,
		int(votedStats.VotedUnitsCount),
		strategy,
	)

	// Build results for each matter
	var results []domain.VoteMatterResult
	for _, dbMatter := range dbMatters {
		matter := domain.DBVotingMatterToResponse(dbMatter)

		// Find tally for this matter
		var tallyData map[string]domain.TallyResult
		for _, tally := range dbTallies {
			if tally.VotingMatterID == matter.ID {
				json.Unmarshal([]byte(tally.TallyData), &tallyData)
				break
			}
		}

		if tallyData == nil {
			tallyData = make(map[string]domain.TallyResult)
		}

		// Calculate statistics for this matter
		var totalVoted float64
		var totalAbstained float64
		var voteResults []domain.VoteResult

		for choice, tally := range tallyData {
			if choice == "abstain" {
				totalAbstained = tally.Weight
			} else {
				totalVoted += tally.Weight
			}

			voteResults = append(voteResults, domain.VoteResult{
				Choice:           choice,
				VoteCount:        tally.Count,
				WeightSum:        tally.Weight,
				Percentage:       tally.Percentage,
				WeightPercentage: tally.Percentage,
			})
		}

		// Determine if this matter passed
		matterResult := domain.VoteMatterResult{
			MatterID:       matter.ID,
			MatterTitle:    matter.Title,
			MatterType:     matter.MatterType,
			VotingConfig:   matter.VotingConfig,
			Votes:          voteResults,
			QuorumInfo:     &quorumInfo,
			Tally:          tallyData,
			TotalVoted:     totalVoted,
			TotalAbstained: totalAbstained,
		}

		// Calculate if passed using quorum service
		matterResult.IsPassed = s.quorumService.CalculateIfPassed(matterResult, matter.VotingConfig, dbGathering)
		if matterResult.IsPassed {
			matterResult.Result = "passed"
		} else {
			matterResult.Result = "failed"
		}

		// Add matter statistics
		var participantCount int
		var totalVotes int
		for _, tally := range tallyData {
			participantCount += tally.Count
			totalVotes += tally.Count
		}

		matterResult.Statistics = domain.MatterStatistics{
			TotalParticipants: participantCount,
			TotalVotes:        totalVotes,
			TotalWeight:       totalVoted + totalAbstained,
			Abstentions:       int(totalAbstained),
			ParticipationRate: 0, // Will be calculated based on gathering stats
		}

		results = append(results, matterResult)
	}

	// Extract area stats
	participatingArea := float64(0)
	if participatingStats.ParticipatingUnitsTotalArea != nil {
		participatingArea = participatingStats.ParticipatingUnitsTotalArea.(float64)
	}

	votedArea := float64(0)
	if votedStats.VotedUnitsTotalArea != nil {
		votedArea = votedStats.VotedUnitsTotalArea.(float64)
	}

	// Build summary
	summary := domain.GatheringSummary{
		QualifiedUnits:      gathering.QualifiedUnitsCount,
		QualifiedWeight:     gathering.QualifiedUnitsTotalPart,
		QualifiedArea:       gathering.QualifiedUnitsTotalArea,
		ParticipatingUnits:  int(participatingStats.ParticipatingUnitsCount),
		ParticipatingWeight: participatingWeight,
		ParticipatingArea:   participatingArea,
		VotedUnits:          int(votedStats.VotedUnitsCount),
		VotedWeight:         votedWeight,
		VotedArea:           votedArea,
		VotingMode:          gathering.VotingMode,
	}

	if gathering.QualifiedUnitsCount > 0 {
		summary.ParticipationRate = (float64(summary.ParticipatingUnits) / float64(gathering.QualifiedUnitsCount)) * 100
	}

	if summary.ParticipatingUnits > 0 {
		summary.VotingCompletionRate = (float64(summary.VotedUnits) / float64(summary.ParticipatingUnits)) * 100
	}

	// Build final results structure
	voteResults := &domain.VoteResults{
		GatheringID: gatheringID,
		Results:     results,
		Summary:     summary,
		GeneratedAt: time.Now().Format(time.RFC3339),
	}

	// Cache the results in database
	err = s.storeResults(ctx, gatheringID, voteResults, &quorumInfo)
	if err != nil {
		log.Printf("[VotingResultsService] Warning: Failed to cache results: %v", err)
		// Don't fail the operation if caching fails
	}

	log.Printf("[VotingResultsService] Successfully computed results for gathering %d", gatheringID)
	return voteResults, nil
}

// GetCachedResults retrieves cached results or computes them if not cached
func (s *VotingResultsService) GetCachedResults(ctx context.Context, gatheringID int64, associationID int64) (*domain.VoteResults, error) {
	log.Printf("[VotingResultsService] Retrieving results for gathering %d", gatheringID)

	// Try to get cached results
	cachedResults, err := s.db.GetVotingResults(ctx, gatheringID)
	if err == nil {
		// Cache hit - parse and return
		log.Printf("[VotingResultsService] Cache hit for gathering %d", gatheringID)
		var results domain.VoteResults
		if err := json.Unmarshal([]byte(cachedResults.ResultsData), &results); err == nil {
			return &results, nil
		}
		log.Printf("[VotingResultsService] Warning: Failed to parse cached results: %v", err)
	}

	// Cache miss or parse error - compute fresh results
	if err != nil && err != sql.ErrNoRows {
		log.Printf("[VotingResultsService] Warning: Error checking cache: %v", err)
	}

	log.Printf("[VotingResultsService] Cache miss for gathering %d, computing fresh results", gatheringID)
	return s.ComputeAndStoreResults(ctx, gatheringID, associationID)
}

// InvalidateResults clears cached results for a gathering
func (s *VotingResultsService) InvalidateResults(ctx context.Context, gatheringID int64) error {
	log.Printf("[VotingResultsService] Invalidating cached results for gathering %d", gatheringID)
	err := s.db.DeleteVotingResults(ctx, gatheringID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to invalidate results: %w", err)
	}
	log.Printf("[VotingResultsService] Successfully invalidated results for gathering %d", gatheringID)
	return nil
}

// storeResults stores computed results in the cache table
func (s *VotingResultsService) storeResults(ctx context.Context, gatheringID int64, results *domain.VoteResults, quorumInfo *domain.QuorumInfo) error {
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Calculate total possible votes
	totalPossibleWeight := results.Summary.QualifiedWeight
	totalPossibleCount := results.Summary.QualifiedUnits
	if quorumInfo.GatheringType != "remote" {
		// For initial/repeated, use participated units
		totalPossibleWeight = results.Summary.ParticipatingWeight
		totalPossibleCount = results.Summary.ParticipatingUnits
	}

	// Try to update existing cache first
	_, err = s.db.UpdateVotingResults(ctx, database.UpdateVotingResultsParams{
		ResultsData:               string(resultsJSON),
		VotingMode:                quorumInfo.VotingMode,
		GatheringType:             quorumInfo.GatheringType,
		TotalPossibleVotesWeight:  totalPossibleWeight,
		TotalPossibleVotesCount:   int64(totalPossibleCount),
		QuorumThresholdPercentage: quorumInfo.RequiredPercentage,
		QuorumMet:                 quorumInfo.Met,
		GatheringID:               gatheringID,
	})

	if err != nil {
		// Update failed, try insert
		_, err = s.db.CreateVotingResults(ctx, database.CreateVotingResultsParams{
			GatheringID:               gatheringID,
			ResultsData:               string(resultsJSON),
			VotingMode:                quorumInfo.VotingMode,
			GatheringType:             quorumInfo.GatheringType,
			TotalPossibleVotesWeight:  totalPossibleWeight,
			TotalPossibleVotesCount:   int64(totalPossibleCount),
			QuorumThresholdPercentage: quorumInfo.RequiredPercentage,
			QuorumMet:                 quorumInfo.Met,
		})
		if err != nil {
			return fmt.Errorf("failed to create cached results: %w", err)
		}
	}

	return nil
}
