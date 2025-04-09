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
    updated_at    = NOW()
WHERE login = ? RETURNING *;

-- name: GetUserByLogin :one
SELECT *
FROM users
where login = ?;