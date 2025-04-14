-- name: CreateUser :one
INSERT INTO users (login, password_hash, topt_secret, is_admin)
VALUES (?,
        ?, ?, ?) RETURNING *;

-- name: DeleteUsers :exec
DELETE
FROM users;

-- name: UpdateUserEmailAndPassword :one
UPDATE users
set password_hash = ?,
    topt_secret   = ?,
    is_admin      = ?,
    updated_at    = datetime()
WHERE login = ? RETURNING *;

-- name: GetUserByLogin :one
SELECT *
FROM users
where login = ?;

-- name: GetUserAssociationsByLogin :many
SELECT users_associations.association_id
FROM users,
     users_associations
WHERE users.id = users_associations.user_id and  users.login = ?;