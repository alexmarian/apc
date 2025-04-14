-- name: GetUnitOwners :many
SELECT owners.id,
       owners.name,
       owners.normalized_name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.first_detected_at,
       owners.created_at,
       owners.updated_at
FROM owners,
     ownerships
WHERE owners.id = ownerships.owner_id
  AND ownerships.unit_id = ?;

--

-- name: GetAssociationOwners :many
SELECT owners.id,
       owners.name,
       owners.normalized_name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.first_detected_at,
       owners.created_at,
       owners.updated_at
FROM owners,
     ownerships,
     units,
     buildings
WHERE owners.id = ownerships.owner_id
  AND ownerships.unit_id = units.id
  AND units.building_id = buildings.id
  AND buildings.association_id = ?;

--

-- name: GetAssociationOwner :one
SELECT owners.id,
       owners.name,
       owners.normalized_name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.first_detected_at,
       owners.created_at,
       owners.updated_at
FROM owners,
     ownerships,
     units,
     buildings
WHERE owners.id = ownerships.owner_id
  AND ownerships.unit_id = units.id
  AND units.building_id = buildings.id
  AND buildings.association_id = ?
  AND owners.id = ?;

--

-- name: UpdateAssociationOwner :exec

UPDATE owners
SET name                  = ?,
    normalized_name       = ?,
    identification_number = ?,
    contact_phone         = ?,
    contact_email         = ?,
    updated_at            = datetime()
WHERE owners.id = ?
  AND EXISTS (SELECT 1
              FROM ownerships
                       JOIN units ON ownerships.unit_id = units.id
                       JOIN buildings ON units.building_id = buildings.id
              WHERE ownerships.owner_id = owners.id
                AND buildings.association_id = ?);

--
