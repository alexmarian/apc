-- name: GetAssociationAccounts :many
SELECT id, number, destination, description, association_id, is_active, created_at, updated_at
FROM accounts
WHERE association_id = ?;

-- name: GetAccount :one
SELECT id, number, destination, description, association_id, is_active, created_at, updated_at
FROM accounts
WHERE id = ? LIMIT 1;

-- name: CreateAccount :one
INSERT INTO accounts (number, destination, description, association_id)
VALUES (?, ?, ?, ?)
    RETURNING id, number, destination, description, association_id, is_active, created_at, updated_at;

-- name: UpdateAccount :one
UPDATE accounts
SET number = ?,
    destination = ?,
    description = ?,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND association_id = ?
    RETURNING id, number, destination, description, association_id, is_active, created_at, updated_at;

-- name: DisableAccount :exec
UPDATE accounts
SET is_active = FALSE,
    updated_at = CURRENT_TIMESTAMP
WHERE id = ? AND association_id = ?;