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

-- name: UseRegistrationToken :execrows
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

-- name: AddRegistrationTokenAssociation :exec
INSERT INTO registration_token_associations (token, association_id)
VALUES (?, ?);

-- name: GetRegistrationTokenAssociationIds :many
SELECT association_id
FROM registration_token_associations
WHERE token = ?;

-- name: ListRegistrationTokenAssociations :many
SELECT registration_token_associations.token, associations.id, associations.name
FROM registration_token_associations
         JOIN associations ON associations.id = registration_token_associations.association_id
ORDER BY associations.name;
