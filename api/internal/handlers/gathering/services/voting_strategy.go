package services

import (
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// VotingStrategy defines the interface for vote counting strategies
type VotingStrategy interface {
	// CalculateVoteWeight calculates the weight/count for a participant's vote
	CalculateVoteWeight(participant domain.GatheringParticipant, units []handlers.Unit) float64

	// CalculateTotalPossibleVotes calculates total possible votes for gathering
	// Returns (weight, count) where:
	// - weight: the total weight/part value
	// - count: the total number of units
	// For Initial/Repeated gatherings, use participated units
	// For Remote gatherings, use all qualified units
	CalculateTotalPossibleVotes(gathering domain.Gathering, qualifiedUnits []handlers.Unit, participatedUnits []handlers.Unit) (weight float64, count int)

	// GetVotingModeName returns the name of this voting mode
	GetVotingModeName() string
}

// GetVotingStrategy returns the appropriate strategy based on voting mode
func GetVotingStrategy(votingMode string) VotingStrategy {
	switch votingMode {
	case "by_weight":
		return &ByWeightStrategy{}
	case "by_unit":
		return &ByUnitStrategy{}
	default:
		// Default fallback to by_weight for backward compatibility
		return &ByWeightStrategy{}
	}
}
