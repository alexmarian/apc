-- name: GetBuildingUnit :one

SELECT * from units where id = ? and building_id = ?;

--

-- name: GetBuildingUnits :many

SELECT * from units where building_id=?;

-- name: GetBuildingUnitsWithOwners :many

SELECT u.id, u.cadastral_number, u.building_id, u.unit_number, u.address, u.entrance, u.area, u.part, u.unit_type, u.floor, u.room_count, u.created_at, u.updated_at,
       COALESCE(GROUP_CONCAT(ow.name, ', '), '') as owner_names
FROM units u
LEFT JOIN ownerships o ON o.unit_id = u.id AND o.association_id = ? AND o.is_active = true
LEFT JOIN owners ow ON ow.id = o.owner_id
WHERE u.building_id = ?
GROUP BY u.id;

--

-- name: UpdateBuildingUnitById :exec

UPDATE units
SET unit_number = ?, address = ?, entrance = ?, unit_type = ?, floor = ?, room_count = ?, updated_at = datetime()
WHERE id = ? and building_id = ?;

--