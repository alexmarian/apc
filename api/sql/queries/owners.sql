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
