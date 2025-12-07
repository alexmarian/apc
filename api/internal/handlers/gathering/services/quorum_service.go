package services

import (
	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// QuorumService handles quorum validation and vote pass/fail determination
type QuorumService struct {
	db *database.Queries
}

// NewQuorumService creates a new QuorumService
func NewQuorumService(db *database.Queries) *QuorumService {
	return &QuorumService{db: db}
}

// CalculateIfPassed determines if a voting matter has passed based on its results
func (s *QuorumService) CalculateIfPassed(result domain.VoteMatterResult, config domain.VotingConfig, gathering database.Gathering) bool {
	// Informative votes don't have pass/fail - they always "pass" for reporting purposes
	if config.RequiredMajority == "informative" {
		return true
	}

	if result.TotalVoted == 0 {
		return false
	}

	// Check quorum first
	totalPossibleWeight := gathering.QualifiedUnitsTotalPart.Float64
	if config.Quorum > 0 && totalPossibleWeight > 0 {
		quorumPercentage := (result.TotalVoted / totalPossibleWeight) * 100
		if quorumPercentage < config.Quorum {
			return false
		}
	}

	// For yes/no votes
	if config.Type == "yes_no" {
		yesVotes := result.Tally["yes"].Weight
		totalValidVotes := result.TotalVoted
		if !config.AllowAbstention && result.TotalAbstained > 0 {
			totalValidVotes += result.TotalAbstained
		}

		percentage := (yesVotes / totalPossibleWeight) * 100

		switch config.RequiredMajority {
		case "simple":
			return percentage > 50
		case "qualified", "supermajority":
			return percentage >= 66.67
		case "unanimous":
			return percentage >= 100
		case "custom":
			return percentage >= config.RequiredMajorityValue
		}
	}

	// For multiple choice, the option with most votes wins if it meets the threshold
	if config.Type == "multiple_choice" || config.Type == "single_choice" {
		var maxWeight float64
		for _, tally := range result.Tally {
			if tally.Weight > maxWeight {
				maxWeight = tally.Weight
			}
		}

		percentage := (maxWeight / result.TotalVoted) * 100

		switch config.RequiredMajority {
		case "simple":
			return percentage > 50
		case "qualified", "supermajority":
			return percentage >= 66.67
		case "unanimous":
			return percentage >= 100
		case "custom":
			return percentage >= config.RequiredMajorityValue
		}
	}

	return false
}

// ValidateGatheringState checks if a gathering is in the expected state
func (s *QuorumService) ValidateGatheringState(gathering database.Gathering, targetStatus string) bool {
	return gathering.Status == targetStatus
}
