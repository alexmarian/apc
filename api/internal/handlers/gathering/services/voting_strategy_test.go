package services

import (
	"testing"

	"github.com/alexmarian/apc/api/internal/handlers"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
)

// TestByWeightStrategy_CalculateVoteWeight tests vote weight calculation for by-weight mode
func TestByWeightStrategy_CalculateVoteWeight(t *testing.T) {
	strategy := &ByWeightStrategy{}

	tests := []struct {
		name           string
		participant    domain.GatheringParticipant
		units          []handlers.Unit
		expectedWeight float64
	}{
		{
			name: "single unit owner",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1},
				UnitsPart: 10.5,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.5},
			},
			expectedWeight: 10.5,
		},
		{
			name: "multi-unit owner with two units",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2},
				UnitsPart: 25.5,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.5},
				{ID: 2, Part: 15.0},
			},
			expectedWeight: 25.5,
		},
		{
			name: "multi-unit owner with three units",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2, 3},
				UnitsPart: 42.75,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.5},
				{ID: 2, Part: 15.0},
				{ID: 3, Part: 17.25},
			},
			expectedWeight: 42.75,
		},
		{
			name: "zero weight unit",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1},
				UnitsPart: 0.0,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 0.0},
			},
			expectedWeight: 0.0,
		},
		{
			name: "very large weight for precision test",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1},
				UnitsPart: 999.999,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 999.999},
			},
			expectedWeight: 999.999,
		},
		{
			name: "fractional weights summing correctly",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2, 3},
				UnitsPart: 33.33,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 11.11},
				{ID: 2, Part: 11.11},
				{ID: 3, Part: 11.11},
			},
			expectedWeight: 33.33,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight := strategy.CalculateVoteWeight(tt.participant, tt.units)
			if weight != tt.expectedWeight {
				t.Errorf("CalculateVoteWeight() = %v, expected %v", weight, tt.expectedWeight)
			}
		})
	}
}

// TestByUnitStrategy_CalculateVoteWeight tests vote weight calculation for by-unit mode
func TestByUnitStrategy_CalculateVoteWeight(t *testing.T) {
	strategy := &ByUnitStrategy{}

	tests := []struct {
		name          string
		participant   domain.GatheringParticipant
		units         []handlers.Unit
		expectedCount float64
	}{
		{
			name: "single unit owner",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1},
				UnitsPart: 10.5,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.5},
			},
			expectedCount: 1.0,
		},
		{
			name: "multi-unit owner with two units",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2},
				UnitsPart: 25.5,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.5},
				{ID: 2, Part: 15.0},
			},
			expectedCount: 2.0,
		},
		{
			name: "multi-unit owner with five units",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2, 3, 4, 5},
				UnitsPart: 75.0,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 15.0},
				{ID: 2, Part: 15.0},
				{ID: 3, Part: 15.0},
				{ID: 4, Part: 15.0},
				{ID: 5, Part: 15.0},
			},
			expectedCount: 5.0,
		},
		{
			name: "owner with 100 units edge case",
			participant: domain.GatheringParticipant{
				UnitsInfo: make([]int64, 100),
				UnitsPart: 1000.0,
			},
			units: func() []handlers.Unit {
				units := make([]handlers.Unit, 100)
				for i := 0; i < 100; i++ {
					units[i] = handlers.Unit{ID: int64(i + 1), Part: 10.0}
				}
				return units
			}(),
			expectedCount: 100.0,
		},
		{
			name: "units with varying weights count equally",
			participant: domain.GatheringParticipant{
				UnitsInfo: []int64{1, 2, 3},
				UnitsPart: 60.0,
			},
			units: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
			},
			expectedCount: 3.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := strategy.CalculateVoteWeight(tt.participant, tt.units)
			if count != tt.expectedCount {
				t.Errorf("CalculateVoteWeight() = %v, expected %v", count, tt.expectedCount)
			}
		})
	}
}

// TestByWeightStrategy_CalculateTotalPossibleVotes tests total vote calculation for by-weight mode
func TestByWeightStrategy_CalculateTotalPossibleVotes(t *testing.T) {
	strategy := &ByWeightStrategy{}

	tests := []struct {
		name              string
		gathering         domain.Gathering
		qualifiedUnits    []handlers.Unit
		participatedUnits []handlers.Unit
		expectedWeight    float64
		expectedCount     int
	}{
		{
			name: "initial gathering uses participated units",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
				{ID: 4, Part: 40.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
			},
			expectedWeight: 30.0,
			expectedCount:  2,
		},
		{
			name: "repeated gathering uses participated units",
			gathering: domain.Gathering{
				GatheringType: "repeated",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
			},
			expectedWeight: 10.0,
			expectedCount:  1,
		},
		{
			name: "remote gathering uses all qualified units",
			gathering: domain.Gathering{
				GatheringType: "remote",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
			},
			expectedWeight: 60.0,
			expectedCount:  3,
		},
		{
			name: "100% participation in initial gathering",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
			},
			expectedWeight: 30.0,
			expectedCount:  2,
		},
		{
			name: "zero participation in initial gathering",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
			},
			participatedUnits: []handlers.Unit{},
			expectedWeight:    0.0,
			expectedCount:     0,
		},
		{
			name: "large gathering with 100 qualified units",
			gathering: domain.Gathering{
				GatheringType: "remote",
			},
			qualifiedUnits: func() []handlers.Unit {
				units := make([]handlers.Unit, 100)
				for i := 0; i < 100; i++ {
					units[i] = handlers.Unit{ID: int64(i + 1), Part: 10.0}
				}
				return units
			}(),
			participatedUnits: []handlers.Unit{},
			expectedWeight:    1000.0,
			expectedCount:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight, count := strategy.CalculateTotalPossibleVotes(tt.gathering, tt.qualifiedUnits, tt.participatedUnits)
			if weight != tt.expectedWeight {
				t.Errorf("CalculateTotalPossibleVotes() weight = %v, expected %v", weight, tt.expectedWeight)
			}
			if count != tt.expectedCount {
				t.Errorf("CalculateTotalPossibleVotes() count = %v, expected %v", count, tt.expectedCount)
			}
		})
	}
}

// TestByUnitStrategy_CalculateTotalPossibleVotes tests total vote calculation for by-unit mode
func TestByUnitStrategy_CalculateTotalPossibleVotes(t *testing.T) {
	strategy := &ByUnitStrategy{}

	tests := []struct {
		name              string
		gathering         domain.Gathering
		qualifiedUnits    []handlers.Unit
		participatedUnits []handlers.Unit
		expectedWeight    float64
		expectedCount     int
	}{
		{
			name: "initial gathering uses participated units",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
				{ID: 4, Part: 40.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
			},
			expectedWeight: 30.0,
			expectedCount:  2,
		},
		{
			name: "repeated gathering uses participated units",
			gathering: domain.Gathering{
				GatheringType: "repeated",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
			},
			expectedWeight: 10.0,
			expectedCount:  1,
		},
		{
			name: "remote gathering uses all qualified units",
			gathering: domain.Gathering{
				GatheringType: "remote",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
				{ID: 2, Part: 20.0},
				{ID: 3, Part: 30.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 10.0},
			},
			expectedWeight: 60.0,
			expectedCount:  3,
		},
		{
			name: "by-unit mode counts units equally regardless of weight",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 5.0},
				{ID: 2, Part: 95.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 5.0},
				{ID: 2, Part: 95.0},
			},
			expectedWeight: 100.0,
			expectedCount:  2,
		},
		{
			name: "single unit in gathering",
			gathering: domain.Gathering{
				GatheringType: "initial",
			},
			qualifiedUnits: []handlers.Unit{
				{ID: 1, Part: 100.0},
			},
			participatedUnits: []handlers.Unit{
				{ID: 1, Part: 100.0},
			},
			expectedWeight: 100.0,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			weight, count := strategy.CalculateTotalPossibleVotes(tt.gathering, tt.qualifiedUnits, tt.participatedUnits)
			if weight != tt.expectedWeight {
				t.Errorf("CalculateTotalPossibleVotes() weight = %v, expected %v", weight, tt.expectedWeight)
			}
			if count != tt.expectedCount {
				t.Errorf("CalculateTotalPossibleVotes() count = %v, expected %v", count, tt.expectedCount)
			}
		})
	}
}

// TestGetVotingStrategy tests the strategy factory function
func TestGetVotingStrategy(t *testing.T) {
	tests := []struct {
		name         string
		votingMode   string
		expectedType string
	}{
		{
			name:         "by_weight mode returns ByWeightStrategy",
			votingMode:   "by_weight",
			expectedType: "by_weight",
		},
		{
			name:         "by_unit mode returns ByUnitStrategy",
			votingMode:   "by_unit",
			expectedType: "by_unit",
		},
		{
			name:         "invalid mode defaults to ByWeightStrategy",
			votingMode:   "invalid",
			expectedType: "by_weight",
		},
		{
			name:         "empty mode defaults to ByWeightStrategy",
			votingMode:   "",
			expectedType: "by_weight",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := GetVotingStrategy(tt.votingMode)
			if strategy.GetVotingModeName() != tt.expectedType {
				t.Errorf("GetVotingStrategy(%q) returned strategy with name %q, expected %q",
					tt.votingMode, strategy.GetVotingModeName(), tt.expectedType)
			}
		})
	}
}

// TestVotingStrategy_GetVotingModeName tests the voting mode name retrieval
func TestVotingStrategy_GetVotingModeName(t *testing.T) {
	tests := []struct {
		name         string
		strategy     VotingStrategy
		expectedName string
	}{
		{
			name:         "ByWeightStrategy returns by_weight",
			strategy:     &ByWeightStrategy{},
			expectedName: "by_weight",
		},
		{
			name:         "ByUnitStrategy returns by_unit",
			strategy:     &ByUnitStrategy{},
			expectedName: "by_unit",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.strategy.GetVotingModeName()
			if name != tt.expectedName {
				t.Errorf("GetVotingModeName() = %q, expected %q", name, tt.expectedName)
			}
		})
	}
}

// TestVotingStrategy_EdgeCases tests edge cases across both strategies
func TestVotingStrategy_EdgeCases(t *testing.T) {
	t.Run("by_weight with nil units slice", func(t *testing.T) {
		strategy := &ByWeightStrategy{}
		participant := domain.GatheringParticipant{
			UnitsInfo: []int64{},
			UnitsPart: 0.0,
		}
		weight := strategy.CalculateVoteWeight(participant, nil)
		if weight != 0.0 {
			t.Errorf("Expected 0.0 for nil units, got %v", weight)
		}
	})

	t.Run("by_unit with nil units slice", func(t *testing.T) {
		strategy := &ByUnitStrategy{}
		participant := domain.GatheringParticipant{
			UnitsInfo: []int64{},
			UnitsPart: 0.0,
		}
		count := strategy.CalculateVoteWeight(participant, nil)
		if count != 0.0 {
			t.Errorf("Expected 0.0 for nil units, got %v", count)
		}
	})

	t.Run("total possible votes with nil qualified units for remote gathering", func(t *testing.T) {
		strategy := &ByWeightStrategy{}
		gathering := domain.Gathering{GatheringType: "remote"}
		weight, count := strategy.CalculateTotalPossibleVotes(gathering, nil, nil)
		if weight != 0.0 || count != 0 {
			t.Errorf("Expected (0.0, 0) for nil units, got (%v, %v)", weight, count)
		}
	})

	t.Run("unknown gathering type defaults to participated units", func(t *testing.T) {
		strategy := &ByWeightStrategy{}
		gathering := domain.Gathering{GatheringType: "unknown"}
		qualifiedUnits := []handlers.Unit{{ID: 1, Part: 100.0}}
		participatedUnits := []handlers.Unit{{ID: 2, Part: 50.0}}

		weight, count := strategy.CalculateTotalPossibleVotes(gathering, qualifiedUnits, participatedUnits)
		if weight != 50.0 || count != 1 {
			t.Errorf("Expected (50.0, 1) for unknown type using participated units, got (%v, %v)", weight, count)
		}
	})
}
