-- name: GetUnitOwnerships :many
SELECT o.*,
       ow.name            as owner_name,
       ow.normalized_name as owner_normalized_name,
       ow.identification_number
FROM ownerships o,
     owners ow
WHERE o.owner_id = ow.id
  AND o.unit_id = ?
  and o.association_id = ?;

--

-- name: GetActiveUnitOwnerships :many
SELECT o.*,
       ow.name            as owner_name,
       ow.normalized_name as owner_normalized_name,
       ow.identification_number
FROM ownerships o,
     owners ow
WHERE o.owner_id = ow.id
  AND o.unit_id = ?
  and o.association_id = ?
  AND is_active = true;
--

-- name: GetUnitOwnership :one
SELECT o.*,
       ow.name            as owner_name,
       ow.normalized_name as owner_normalized_name,
       ow.identification_number
FROM ownerships o,
     owners ow
WHERE o.owner_id = ow.id
  AND o.unit_id = ?
  AND o.association_id = ?
  AND o.id = ?;

--
    
-- name: DisableActiveVoting :exec
UPDATE ownerships
SET is_voting = false
WHERE unit_id = ?
  AND association_id = ?
  AND is_voting = true;
--

-- name: SetVoting :exec
UPDATE ownerships
SET is_voting = true
WHERE id = ?
  AND unit_id = ?
  AND association_id = ?;
--

-- name: GetOwnership :one
SELECT *
FROM ownerships
WHERE id = ? LIMIT 1;

--