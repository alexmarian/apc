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

// CalculateQuorum calculates quorum information based on gathering type and voting mode
func (s *QuorumService) CalculateQuorum(gathering domain.Gathering, participatedWeight float64, participatedCount int, votedWeight float64, votedCount int, strategy VotingStrategy) domain.QuorumInfo {
	// Step 1: Determine threshold percentage based on gathering type
	var thresholdPercent float64
	switch gathering.GatheringType {
	case "initial":
		thresholdPercent = 50.0
	case "repeated":
		thresholdPercent = 25.0
	case "remote":
		thresholdPercent = 100.0
	default:
		thresholdPercent = 50.0 // Default to initial
	}

	// Step 2: Calculate total possible votes based on gathering type and voting mode
	var totalPossibleWeight float64
	var totalPossibleCount int
	var achieved float64

	if gathering.GatheringType == "remote" {
		// Remote gatherings: total possible = all qualified units
		totalPossibleWeight = gathering.QualifiedUnitsTotalPart
		totalPossibleCount = gathering.QualifiedUnitsCount
	} else {
		// Initial/Repeated gatherings: total possible = participated units
		totalPossibleWeight = participatedWeight
		totalPossibleCount = participatedCount
	}

	// Step 3: Calculate achieved participation based on voting mode
	if gathering.VotingMode == "by_unit" {
		achieved = float64(votedCount)
	} else {
		// by_weight (default)
		achieved = votedWeight
	}

	// Step 4: Calculate required amount based on voting mode
	var required float64
	var totalPossible float64
	if gathering.VotingMode == "by_unit" {
		totalPossible = float64(totalPossibleCount)
		required = (totalPossible * thresholdPercent) / 100
	} else {
		// by_weight (default)
		totalPossible = totalPossibleWeight
		required = (totalPossible * thresholdPercent) / 100
	}

	// Step 5: Calculate achieved percentage
	var achievedPercentage float64
	if totalPossible > 0 {
		achievedPercentage = (achieved / totalPossible) * 100
	}

	// Step 6: Determine if quorum is met
	met := achieved >= required

	return domain.QuorumInfo{
		Required:           required,
		Achieved:           achieved,
		RequiredPercentage: thresholdPercent,
		AchievedPercentage: achievedPercentage,
		Met:                met,
		VotingMode:         gathering.VotingMode,
		GatheringType:      gathering.GatheringType,
	}
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
