-- name: GetBuildingUnit :one

SELECT * from units where id = ? and building_id = ?;

--

-- name: GetBuildingUnits :many

SELECT * from units where building_id=?;

--

-- name: UpdateBuildingUnitById :exec

UPDATE units
SET unit_number = ?, address = ?, entrance = ?, unit_type = ?, floor = ?, room_count = ?, updated_at = datetime()
WHERE id = ? and building_id = ?;

--