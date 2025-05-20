-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, login, expires_at)
VALUES (?,
        datetime('now'),
        datetime('now'),
        ?,
        datetime('now', '+1 days'));

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
set revoked_at = datetime('now'),
    updated_at = datetime('now')
where token = ?;

-- name: GetValidRefreshToken :one
SELECT *
FROM refresh_tokens
where token = ?
  AND revoked_at IS NULL
  AND expires_at > datetime('now');