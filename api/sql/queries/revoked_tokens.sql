-- name: RevokeToken :exec
INSERT INTO revoked_tokens (jti, expires_at) VALUES (?, ?);

-- name: IsTokenRevoked :one
SELECT EXISTS(
    SELECT 1 FROM revoked_tokens
    WHERE jti = ? AND expires_at > datetime('now')
) AS revoked;

-- name: PurgeExpiredRevokedTokens :exec
DELETE FROM revoked_tokens WHERE expires_at <= datetime('now');
