-- name: GetAssociations :one

SELECT * from associations where id = ?;

--

-- name: GetAssociationsFromList :many

SELECT * from associations where id in (sqlc.slice('association_ids'));

--

-- name: ListAssociations :many

SELECT * from associations ORDER BY id;

--
-- name: CreateAssociation :one

INSERT INTO associations (name, address, administrator)
VALUES (?, ?, ?)
RETURNING *;

--

-- name: UpdateAssociation :one

UPDATE associations
SET name          = ?,
    address       = ?,
    administrator = ?,
    updated_at    = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING *;

--

-- name: DeleteAssociationUsers :exec

DELETE FROM users_associations WHERE association_id = ?;

--

-- name: AddAssociationUser :exec

INSERT INTO users_associations (user_id, association_id)
VALUES (?, ?);

--

-- name: ListAssociationAdministrators :many

SELECT users_associations.association_id, users.id, users.login
FROM users_associations
         JOIN users ON users.id = users_associations.user_id
ORDER BY users.login;

--
