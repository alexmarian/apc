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

-- name: GetOwnership :one
SELECT *
FROM ownerships
WHERE id = ? LIMIT 1;

--