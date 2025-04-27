-- name: CreateRegistrationToken :one
INSERT INTO registration_tokens (
    token,
    created_by,
    expires_at,
    description,
    is_admin
) VALUES (?, ?, ?, ?, ?)
    RETURNING *;

-- name: GetValidRegistrationToken :one
SELECT *
FROM registration_tokens
WHERE token = ?
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > datetime('now')
    LIMIT 1;

-- name: UseRegistrationToken :exec
UPDATE registration_tokens
SET used_at = datetime('now'),
    used_by = ?
WHERE token = ?
  AND used_at IS NULL
  AND revoked_at IS NULL
  AND expires_at > datetime('now');

-- name: RevokeRegistrationToken :exec
UPDATE registration_tokens
SET revoked_at = datetime('now'),
    revoked_by = ?
WHERE token = ?
  AND used_at IS NULL
  AND revoked_at IS NULL;

-- name: GetAllRegistrationTokens :many
SELECT *
FROM registration_tokens
ORDER BY created_at DESC;

-- name: GetRegistrationTokensStatus :many
SELECT
    token,
    created_by,
    created_at,
    expires_at,
    CASE
        WHEN used_at IS NOT NULL THEN 'Used'
        WHEN revoked_at IS NOT NULL THEN 'Revoked'
        WHEN expires_at < datetime('now') THEN 'Expired'
        ELSE 'Valid'
        END as status,
    used_at,
    used_by,
    revoked_at,
    revoked_by,
    description,
    is_admin
FROM registration_tokens
ORDER BY created_at DESC;