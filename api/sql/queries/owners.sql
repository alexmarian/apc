-- name: GetUnitOwners :many
SELECT owners.id,
       owners.name,
       owners.normalized_name,
       owners.identification_number,
       owners.contact_phone,
       owners.contact_email,
       owners.first_detected_at,
       owners.association_id,
       owners.created_at,
       owners.updated_at
FROM owners,
     ownerships
WHERE owners.id = ownerships.owner_id
  AND ownerships.unit_id = ? and ownerships.association_id = ?;

--

-- name: GetAssociationOwners :many
SELECT *
FROM owners
WHERE owners.association_id = ?;

--

-- name: GetAssociationOwner :one
SELECT *
FROM owners
WHERE owners.id = ? and owners.association_id=?;

--
    
-- name: GetOwnerById :one
SELECT * FROM owners
WHERE id = ? LIMIT 1;
--

-- name: UpdateAssociationOwner :exec

UPDATE owners
SET name                  = ?,
    normalized_name       = ?,
    identification_number = ?,
    contact_phone         = ?,
    contact_email         = ?,
    updated_at            = datetime()
WHERE id = ? AND association_id = ?;

--

-- name: CreateOwner :one
INSERT INTO owners (
    name, normalized_name, identification_number,
    contact_phone, contact_email, association_id
) VALUES (?, ?, ?, ?, ?, ?)
    RETURNING *;
--

-- name: GetActiveUnitOwnerships :many
SELECT * FROM ownerships
WHERE unit_id = ? AND is_active = true;
--

-- name: DeactivateOwnership :exec
UPDATE ownerships
SET is_active = false, end_date = ?, updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

--

-- name: CreateOwnership :one
INSERT INTO ownerships (
    unit_id, owner_id, association_id,
    start_date, end_date, is_active,
    registration_document, registration_date
) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
    RETURNING *;
--

-- name: GetOwnerUnitsWithDetails :many
SELECT
    u.id as unit_id,
    u.unit_number,
    u.area,
    u.part,
    u.unit_type,
    b.name as building_name,
    b.address as building_address,
    o2.id as co_owner_id,
    o2.name as co_owner_name,
    o2.normalized_name as co_owner_normalized_name,
    o2.identification_number as co_owner_identification_number,
    o2.contact_phone as co_owner_contact_phone,
    o2.contact_email as co_owner_contact_email
FROM ownerships o
         JOIN units u ON o.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
         LEFT JOIN ownerships o_co ON o.unit_id = o_co.unit_id AND o_co.is_active = true AND o_co.owner_id != o.owner_id
LEFT JOIN owners o2 ON o_co.owner_id = o2.id
WHERE o.owner_id = ? AND o.association_id = ? AND o.is_active = true;

--

-- name: GetOwnerUnitsWithDetailsForReport :many
SELECT
    o.id as owner_id,
    o.name as owner_name,
    o.normalized_name as owner_normalized_name,
    o.identification_number as owner_identification_number,
    o.contact_phone as owner_contact_phone,
    o.contact_email as owner_contact_email,
    o.first_detected_at as owner_first_detected_at,
    o.created_at as owner_created_at,
    o.updated_at as owner_updated_at,
    u.id as unit_id,
    u.unit_number,
    u.area,
    u.part,
    u.unit_type,
    b.name as building_name,
    b.address as building_address,
    o2.id as co_owner_id,
    o2.name as co_owner_name,
    o2.normalized_name as co_owner_normalized_name,
    o2.identification_number as co_owner_identification_number,
    o2.contact_phone as co_owner_contact_phone,
    o2.contact_email as co_owner_contact_email
FROM owners o
         JOIN ownerships os ON o.id = os.owner_id
         JOIN units u ON os.unit_id = u.id
         JOIN buildings b ON u.building_id = b.id
         LEFT JOIN ownerships os2 ON os.unit_id = os2.unit_id AND os2.is_active = true AND os2.owner_id != o.id
LEFT JOIN owners o2 ON os2.owner_id = o2.id
WHERE
    o.association_id = ? AND
    os.is_active = true AND
    (? = 0 OR o.id = ?) -- This filters by owner_id if provided, otherwise returns all
ORDER BY o.id, u.id;