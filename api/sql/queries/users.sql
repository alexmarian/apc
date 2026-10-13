-- name: CreateUser :one
INSERT INTO users (login, password_hash, topt_secret, is_admin, totp_confirmed_at)
VALUES (?,
        ?, ?, ?, ?) RETURNING *;

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

-- name: GetUsersByIds :many
SELECT id, login, is_admin, disabled_at
FROM users
WHERE id IN (sqlc.slice('ids'));

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = ?;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY login;

-- name: ListUserAssociationLinks :many
SELECT users_associations.user_id, associations.id, associations.name
FROM users_associations
         JOIN associations ON associations.id = users_associations.association_id
ORDER BY associations.name;

-- name: DeleteUserAssociations :exec
DELETE
FROM users_associations
WHERE user_id = ?;

-- name: ConfirmUserTotp :exec
UPDATE users
SET totp_confirmed_at = datetime('now'),
    updated_at        = datetime('now')
WHERE id = ?;

-- name: SetUserTotpSecret :exec
-- A new secret always needs confirming again.
UPDATE users
SET topt_secret       = ?,
    totp_confirmed_at = NULL,
    updated_at        = datetime('now')
WHERE id = ?;

-- name: SetUserPassword :exec
UPDATE users
SET password_hash = ?,
    updated_at    = datetime('now')
WHERE id = ?;

-- name: SetUserAdmin :exec
UPDATE users
SET is_admin   = ?,
    updated_at = datetime('now')
WHERE id = ?;

-- name: SetUserDisabled :exec
UPDATE users
SET disabled_at = ?,
    disabled_by = ?,
    updated_at  = datetime('now')
WHERE id = ?;

-- name: CountActiveAdmins :one
SELECT COUNT(*)
FROM users
WHERE is_admin = TRUE
  AND disabled_at IS NULL;
