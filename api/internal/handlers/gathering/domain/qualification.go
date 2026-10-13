package domain

import (
	"database/sql"
	"encoding/json"

	"github.com/alexmarian/apc/api/internal/database"
)

// Qualification is which units may vote in a gathering: those matching every
// non-empty list (unit types AND floors AND entrances AND buildings); an
// empty list doesn't restrict. Several buildings make one common decision —
// their units vote together, with one quorum.
type Qualification struct {
	UnitTypes []string
	Floors    []int64
	Entrances []int64
	Buildings []int64
}

func listOf[T any](s sql.NullString) []T {
	var v []T
	if s.Valid {
		json.Unmarshal([]byte(s.String), &v)
	}
	return v
}

// QualificationOf reads a gathering's stored qualification.
func QualificationOf(g database.Gathering) Qualification {
	return Qualification{
		UnitTypes: listOf[string](g.QualificationUnitTypes),
		Floors:    listOf[int64](g.QualificationFloors),
		Entrances: listOf[int64](g.QualificationEntrances),
		Buildings: listOf[int64](g.QualificationBuildings),
	}
}

// Every "has…" flag is set explicitly: left nil, the query's `false = ?`
// would be NULL and exclude every unit.

func (q Qualification) QualifiedUnitsParams(associationID int64) database.GetQualifiedUnitsParams {
	return database.GetQualifiedUnitsParams{
		AssociationID: associationID,
		Column2:       len(q.UnitTypes) > 0,
		UnitTypes:     q.UnitTypes,
		Column4:       len(q.Floors) > 0,
		UnitFloors:    q.Floors,
		Column6:       len(q.Entrances) > 0,
		UnitEntrances: q.Entrances,
		Column8:       len(q.Buildings) > 0,
		BuildingIds:   q.Buildings,
	}
}

func (q Qualification) OwnerUnitsParams(associationID int64) database.GetActiveOwnerUnitsForGatheringParams {
	return database.GetActiveOwnerUnitsForGatheringParams{
		AssociationID: associationID,
		Column2:       len(q.UnitTypes) > 0,
		UnitTypes:     q.UnitTypes,
		Column4:       len(q.Floors) > 0,
		UnitFloors:    q.Floors,
		Column6:       len(q.Entrances) > 0,
		UnitEntrances: q.Entrances,
		Column8:       len(q.Buildings) > 0,
		BuildingIds:   q.Buildings,
	}
}

func (q Qualification) NonParticipatingParams(associationID, gatheringID int64) database.GetNonParticipatingOwnersParams {
	return database.GetNonParticipatingOwnersParams{
		AssociationID:   associationID,
		AssociationID_2: associationID,
		Column3:         len(q.UnitTypes) > 0,
		UnitTypes:       q.UnitTypes,
		Column5:         len(q.Floors) > 0,
		UnitFloors:      q.Floors,
		Column7:         len(q.Entrances) > 0,
		UnitEntrances:   q.Entrances,
		Column9:         len(q.Buildings) > 0,
		BuildingIds:     q.Buildings,
		GatheringID:     gatheringID,
	}
}
