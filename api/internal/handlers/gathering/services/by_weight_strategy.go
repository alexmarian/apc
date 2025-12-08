package services

import (
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// ByWeightStrategy implements voting by weight (combined unit weights per owner)
type ByWeightStrategy struct{}

// CalculateVoteWeight returns the sum of all unit weights for a participant
// In by-weight mode, a participant's vote counts as their total combined weight
func (s *ByWeightStrategy) CalculateVoteWeight(participant domain.GatheringParticipant, units []handlers.Unit) float64 {
	// For by-weight voting, the participant's UnitsPart already contains
	// the sum of all their units' weights
	return participant.UnitsPart
}

// CalculateTotalPossibleVotes calculates the total possible votes based on gathering type
// Returns (weight, count) where weight is the sum of unit weights
func (s *ByWeightStrategy) CalculateTotalPossibleVotes(gathering domain.Gathering, qualifiedUnits []handlers.Unit, participatedUnits []handlers.Unit) (weight float64, count int) {
	var unitsToCount []handlers.Unit

	// Determine which units to count based on gathering type
	switch gathering.GatheringType {
	case "initial", "repeated":
		// For Initial/Repeated gatherings, use participated units
		// Total possible votes = sum of weights of units whose owners participated
		unitsToCount = participatedUnits
	case "remote":
		// For Remote gatherings, use ALL qualified units
		// Total possible votes = sum of weights of all qualified units
		unitsToCount = qualifiedUnits
	default:
		// Fallback to participated units for unknown types
		unitsToCount = participatedUnits
	}

	// Calculate total weight
	totalWeight := 0.0
	for _, unit := range unitsToCount {
		totalWeight += unit.Part
	}

	return totalWeight, len(unitsToCount)
}

// GetVotingModeName returns the name of this voting mode
func (s *ByWeightStrategy) GetVotingModeName() string {
	return "by_weight"
}
