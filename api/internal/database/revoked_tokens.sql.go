package database

import (
	"context"
	"time"
)

const revokeToken = `
INSERT INTO revoked_tokens (jti, expires_at) VALUES (?, ?)
`

func (q *Queries) RevokeToken(ctx context.Context, jti string, expiresAt time.Time) error {
	_, err := q.db.ExecContext(ctx, revokeToken, jti, expiresAt)
	return err
}

const isTokenRevoked = `
SELECT EXISTS(
    SELECT 1 FROM revoked_tokens
    WHERE jti = ? AND expires_at > datetime('now')
) AS revoked
`

func (q *Queries) IsTokenRevoked(ctx context.Context, jti string) (bool, error) {
	var revoked bool
	err := q.db.QueryRowContext(ctx, isTokenRevoked, jti).Scan(&revoked)
	return revoked, err
}

const purgeExpiredRevokedTokens = `
DELETE FROM revoked_tokens WHERE expires_at <= datetime('now')
`

func (q *Queries) PurgeExpiredRevokedTokens(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, purgeExpiredRevokedTokens)
	return err
}
