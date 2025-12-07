package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"math"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// StatsService handles gathering statistics calculations
type StatsService struct {
	db *database.Queries
}

// NewStatsService creates a new StatsService
func NewStatsService(db *database.Queries) *StatsService {
	return &StatsService{db: db}
}

// RoundTo3Decimals rounds a float to 3 decimal places
func RoundTo3Decimals(value float64) float64 {
	return math.Round(value*1000) / 1000
}

// UpdateGatheringStats calculates and updates gathering statistics
func (s *StatsService) UpdateGatheringStats(gatheringID, associationID int64) (int, float64, float64) {
	ctx := context.Background()

	// Get gathering details
	gathering, err := s.db.GetGathering(ctx, database.GetGatheringParams{
		ID:            gatheringID,
		AssociationID: associationID,
	})
	if err != nil {
		return 0, 0.0, 0.0
	}

	// Parse qualification rules
	var unitTypes []string
	var floors []int64
	var entrances []int64

	if gathering.QualificationUnitTypes.Valid {
		json.Unmarshal([]byte(gathering.QualificationUnitTypes.String), &unitTypes)
	}
	if gathering.QualificationFloors.Valid {
		json.Unmarshal([]byte(gathering.QualificationFloors.String), &floors)
	}
	if gathering.QualificationEntrances.Valid {
		json.Unmarshal([]byte(gathering.QualificationEntrances.String), &entrances)
	}

	// Get qualified units
	units, err := s.db.GetQualifiedUnits(ctx, database.GetQualifiedUnitsParams{
		AssociationID: associationID,
		Column2:       len(unitTypes) > 0,
		UnitTypes:     unitTypes,
		Column4:       len(floors) > 0,
		UnitFloors:    floors,
		Column6:       len(entrances) > 0,
		UnitEntrances: entrances,
	})

	if err != nil {
		return 0, 0.0, 0.0
	}

	// Calculate totals
	qualifiedCount := len(units)
	qualifiedTotalPart := 0.0
	qualifiedTotalArea := 0.0
	for _, u := range units {
		qualifiedTotalPart += u.Part
		qualifiedTotalArea += u.Area
	}

	// Update stats
	s.db.UpdateGatheringStats(ctx, database.UpdateGatheringStatsParams{
		QualifiedUnitsCount:     sql.NullInt64{Int64: int64(qualifiedCount), Valid: true},
		QualifiedUnitsTotalPart: sql.NullFloat64{Float64: qualifiedTotalPart, Valid: true},
		QualifiedUnitsTotalArea: sql.NullFloat64{Float64: qualifiedTotalArea, Valid: true},
		ID:                      gatheringID,
	})
	return qualifiedCount, qualifiedTotalPart, qualifiedTotalArea
}

// UpdateGatheringParticipationStats updates participation statistics for a gathering
func (s *StatsService) UpdateGatheringParticipationStats(gatheringID, associationID int64) {
	ctx := context.Background()

	// Get participating units stats (unit-based, not participant-based)
	stats, err := s.db.GetParticipatingUnitsStats(ctx, gatheringID)
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting participating units stats", zap.Error(err))
		return
	}

	// Type assert interface{} to float64
	participatingPart, _ := stats.ParticipatingUnitsTotalPart.(float64)
	participatingArea, _ := stats.ParticipatingUnitsTotalArea.(float64)

	// Update stats
	s.db.UpdateParticipationStats(ctx, database.UpdateParticipationStatsParams{
		ParticipatingUnitsCount:     sql.NullInt64{Int64: stats.ParticipatingUnitsCount, Valid: true},
		ParticipatingUnitsTotalPart: sql.NullFloat64{Float64: participatingPart, Valid: true},
		ParticipatingUnitsTotalArea: sql.NullFloat64{Float64: participatingArea, Valid: true},
		ID:                          gatheringID,
	})
}

// CalculateFinalResults finalizes all tallies and participation stats
func (s *StatsService) CalculateFinalResults(gatheringID, associationID int64, tallyService *TallyService) {
	// Update participation stats with final unit counts
	s.UpdateGatheringParticipationStats(gatheringID, associationID)

	// Update vote tallies
	tallyService.UpdateVoteTallies(gatheringID, -1)

	logging.Logger.Log(zap.InfoLevel, "Final results calculated",
		zap.Int64("gathering_id", gatheringID),
		zap.Int64("association_id", associationID))
}
