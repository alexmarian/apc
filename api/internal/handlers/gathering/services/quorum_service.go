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

// CalculateQuorum calculates quorum information based on gathering type and voting mode.
// Quorum is always expressed as a fraction of qualified units (by count or by part),
// never of checked-in participants.
func (s *QuorumService) CalculateQuorum(gathering domain.Gathering, _ float64, _ int, votedWeight float64, votedCount int, _ VotingStrategy) domain.QuorumInfo {
	// Threshold percentage by gathering type
	var thresholdPercent float64
	switch gathering.GatheringType {
	case "initial":
		thresholdPercent = 50.0
	case "repeated":
		thresholdPercent = 25.0
	case "remote":
		thresholdPercent = 100.0
	default:
		thresholdPercent = 50.0
	}

	// Denominator is always total qualified units
	totalPossibleWeight := gathering.QualifiedUnitsTotalPart
	totalPossibleCount := gathering.QualifiedUnitsCount

	var achieved, required, totalPossible float64
	if gathering.VotingMode == "by_unit" {
		totalPossible = float64(totalPossibleCount)
		achieved = float64(votedCount)
	} else {
		totalPossible = totalPossibleWeight
		achieved = votedWeight
	}
	required = totalPossible * thresholdPercent / 100

	var achievedPercentage float64
	if totalPossible > 0 {
		achievedPercentage = achieved / totalPossible * 100
	}

	return domain.QuorumInfo{
		Required:           required,
		Achieved:           achieved,
		RequiredPercentage: thresholdPercent,
		AchievedPercentage: achievedPercentage,
		Met:                achieved >= required,
		VotingMode:         gathering.VotingMode,
		GatheringType:      gathering.GatheringType,
	}
}

// CalculateIfPassed determines if a voting matter has passed based on its results.
// Informative matters are handled by the caller via VotingMatter.IsInformative.
func (s *QuorumService) CalculateIfPassed(result domain.VoteMatterResult, config domain.VotingConfig, gathering database.Gathering) bool {
	// Poll (sondaj) matters are always accepted regardless of participation or quorum
	if result.MatterType == "poll" {
		return true
	}

	if result.TotalVoted == 0 {
		return false
	}

	// Gathering-level quorum must be met before any matter can pass
	if result.QuorumInfo != nil && !result.QuorumInfo.Met {
		return false
	}

	qualifiedWeight := gathering.QualifiedUnitsTotalPart.Float64
	qualifiedCount := float64(gathering.QualifiedUnitsCount.Int64)

	// Quorum check: votes cast vs qualified units (respects voting_mode)
	if config.Quorum > 0 {
		var quorumDenominator float64
		if gathering.VotingMode == "by_unit" {
			quorumDenominator = qualifiedCount
		} else {
			quorumDenominator = qualifiedWeight
		}
		if quorumDenominator > 0 {
			quorumPct := result.TotalVoted / quorumDenominator * 100
			if quorumPct < config.Quorum {
				return false
			}
		}
	}

	// Determine the winning vote weight and the denominator for majority calculation
	var winningWeight float64
	if config.Type == "yes_no" {
		winningWeight = result.Tally["yes"].Weight
	} else {
		for _, tally := range result.Tally {
			if tally.Weight > winningWeight {
				winningWeight = tally.Weight
			}
		}
	}

	// absolute and absolute_two_thirds use all qualified voters as denominator
	var denominator float64
	switch config.RequiredMajority {
	case "absolute", "absolute_two_thirds":
		if gathering.VotingMode == "by_unit" {
			denominator = qualifiedCount
		} else {
			denominator = qualifiedWeight
		}
	default:
		// simple, qualified, unanimous: denominator is votes cast (excluding abstentions)
		denominator = result.TotalVoted
	}

	if denominator == 0 {
		return false
	}
	percentage := winningWeight / denominator * 100

	switch config.RequiredMajority {
	case "simple":
		return percentage > 50
	case "absolute":
		return percentage > 50
	case "absolute_two_thirds":
		return percentage >= 66.67
	case "qualified":
		threshold := config.RequiredMajorityValue
		if threshold == 0 {
			threshold = 66.67
		}
		return percentage >= threshold
	case "unanimous":
		return percentage >= 100
	}

	return false
}

// ValidateGatheringState checks if a gathering is in the expected state
func (s *QuorumService) ValidateGatheringState(gathering database.Gathering, targetStatus string) bool {
	return gathering.Status == targetStatus
}
