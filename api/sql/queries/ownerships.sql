-- name: GetUnitOwnerships :many
SELECT *
FROM  ownerships
WHERE ownerships.unit_id = ? and ownerships.association_id = ?;

--