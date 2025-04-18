// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: ownerships.sql

package database

import (
	"context"
)

const getUnitOwnerships = `-- name: GetUnitOwnerships :many
SELECT id, unit_id, owner_id, association_id, start_date, end_date, is_active, registration_document, registration_date, created_at, updated_at
FROM  ownerships
WHERE ownerships.unit_id = ? and ownerships.association_id = ?
`

type GetUnitOwnershipsParams struct {
	UnitID        int64
	AssociationID int64
}

func (q *Queries) GetUnitOwnerships(ctx context.Context, arg GetUnitOwnershipsParams) ([]Ownership, error) {
	rows, err := q.db.QueryContext(ctx, getUnitOwnerships, arg.UnitID, arg.AssociationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Ownership
	for rows.Next() {
		var i Ownership
		if err := rows.Scan(
			&i.ID,
			&i.UnitID,
			&i.OwnerID,
			&i.AssociationID,
			&i.StartDate,
			&i.EndDate,
			&i.IsActive,
			&i.RegistrationDocument,
			&i.RegistrationDate,
			&i.CreatedAt,
			&i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
