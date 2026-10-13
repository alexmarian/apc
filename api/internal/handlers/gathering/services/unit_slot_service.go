package services

import (
	"context"

	"github.com/alexmarian/apc/api/internal/database"
	"github.com/alexmarian/apc/api/internal/handlers/gathering/domain"
	"github.com/alexmarian/apc/api/internal/logging"
	"go.uber.org/zap"
)

// UnitSlotService handles unit slot management for gatherings
type UnitSlotService struct {
	db *database.Queries
}

// NewUnitSlotService creates a new UnitSlotService
func NewUnitSlotService(db *database.Queries) *UnitSlotService {
	return &UnitSlotService{db: db}
}

// SyncUnitsSlots creates unit slots for all qualified units in a gathering
func (s *UnitSlotService) SyncUnitsSlots(ctx context.Context, associationID int64, gatheringID int64, q domain.Qualification) error {
	units, err := s.db.GetQualifiedUnits(ctx, q.QualifiedUnitsParams(associationID))
	if err != nil {
		logging.Logger.Log(zap.WarnLevel, "Error getting qualified units", zap.Error(err))
		return err
	}

	for _, unit := range units {
		_, err := s.db.CreateUnitSlot(ctx, database.CreateUnitSlotParams{
			GatheringID: gatheringID,
			UnitID:      unit.ID,
		})
		if err != nil {
			logging.Logger.Log(zap.WarnLevel, "Error creating unit slot", zap.Error(err))
			return err
		}
	}
	return nil
}
