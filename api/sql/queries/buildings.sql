-- name: GetAssociationBuilding :one

SELECT * from buildings where id = ? and association_id = ?;

--

-- name: GetAssociationBuildings :many

SELECT * from buildings where association_id=?;

--
-- name: CreateBuilding :one

INSERT INTO buildings (name, address, cadastral_number, total_area, association_id)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

--

-- name: UpdateBuilding :exec

UPDATE buildings
SET name             = ?,
    address          = ?,
    cadastral_number = ?,
    total_area       = ?,
    updated_at       = datetime('now')
WHERE id = ?
  AND association_id = ?;

--

-- name: DeleteBuilding :exec

DELETE FROM buildings WHERE id = ? AND association_id = ?;

--

-- name: CountBuildingUnits :one

SELECT COUNT(*) FROM units WHERE building_id = ?;

--

-- name: GetBuildingShareTotal :one

SELECT CAST(COALESCE(SUM(part), 0) AS REAL) AS total FROM units WHERE building_id = ?;

--
