-- name: GetUnitOwnerships :many
SELECT *
FROM  ownerships
WHERE ownerships.unit_id = ? and ownerships.association_id = ?;

--

-- name: GetOwnership :one
SELECT * FROM ownerships
WHERE id = ? LIMIT 1;

--