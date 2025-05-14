-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (token,
                                   login,
                                   created_at,
                                   expires_at)
VALUES (?, ?, datetime('now'), ?);
--
-- name: GetValidPasswordResetToken :one
SELECT token, login, created_at, expires_at, used_at
FROM password_reset_tokens
WHERE token = ?
  AND used_at IS NULL
  AND expires_at > datetime('now') LIMIT 1;
--
-- name: UsePasswordResetToken :exec
UPDATE password_reset_tokens
SET used_at = datetime('now')
WHERE token = ?;
--