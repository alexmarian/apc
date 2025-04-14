-- name: GetAssociationBuilding :one

SELECT * from buildings where id = ? and association_id = ?;

--

-- name: GetAssociationBuildings :many

SELECT * from buildings where association_id=?;

--