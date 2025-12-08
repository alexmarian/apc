package services

import (
	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// ByUnitStrategy implements voting by unit (each unit = 1 vote)
type ByUnitStrategy struct{}

// CalculateVoteWeight returns the count of units for a participant
// In by-unit mode, each unit owned by the participant counts as 1 vote
func (s *ByUnitStrategy) CalculateVoteWeight(participant domain.GatheringParticipant, units []handlers.Unit) float64 {
	// For by-unit voting, each unit counts as 1 vote
	// Return the count of units as a float64 for consistency with the interface
	return float64(len(participant.UnitsInfo))
}

// CalculateTotalPossibleVotes calculates the total possible votes based on gathering type
// Returns (weight, count) where count is the number of units
func (s *ByUnitStrategy) CalculateTotalPossibleVotes(gathering domain.Gathering, qualifiedUnits []handlers.Unit, participatedUnits []handlers.Unit) (weight float64, count int) {
	var unitsToCount []handlers.Unit

	// Determine which units to count based on gathering type
	switch gathering.GatheringType {
	case "initial", "repeated":
		// For Initial/Repeated gatherings, use participated units
		// Total possible votes = count of units whose owners participated
		unitsToCount = participatedUnits
	case "remote":
		// For Remote gatherings, use ALL qualified units
		// Total possible votes = count of all qualified units
		unitsToCount = qualifiedUnits
	default:
		// Fallback to participated units for unknown types
		unitsToCount = participatedUnits
	}

	// For by-unit mode, we still need to calculate the total weight for reporting purposes
	// But the primary metric is the count
	totalWeight := 0.0
	for _, unit := range unitsToCount {
		totalWeight += unit.Part
	}

	// Return both weight (for reporting) and count (primary metric for by-unit mode)
	return totalWeight, len(unitsToCount)
}

// GetVotingModeName returns the name of this voting mode
func (s *ByUnitStrategy) GetVotingModeName() string {
	return "by_unit"
}
