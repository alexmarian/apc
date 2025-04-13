-- name: GetAssociations :one

SELECT * from associations where id = ?;

--

-- name: GetAssociationsFromList :many

SELECT * from associations where id in (sqlc.slice('association_ids'));

--