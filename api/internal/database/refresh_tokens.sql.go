// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_tokens.sql

package database

import (
	"context"
)

const createRefreshToken = `-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (token, created_at, updated_at, login, expires_at)
VALUES (?,
        datetime('now'),
        datetime('now'),
        ?,
        datetime('now', '+1 days'))
`

type CreateRefreshTokenParams struct {
	Token string
	Login string
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, createRefreshToken, arg.Token, arg.Login)
	return err
}

const getValidRefreshToken = `-- name: GetValidRefreshToken :one
SELECT token, created_at, updated_at, login, expires_at, revoked_at
FROM refresh_tokens
where token = ?
  AND revoked_at IS NULL
  AND expires_at > datetime('now')
`

func (q *Queries) GetValidRefreshToken(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getValidRefreshToken, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Login,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

const revokeRefreshToken = `-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
set revoked_at = datetime('now'),
    updated_at = datetime('now')
where token = ?
`

func (q *Queries) RevokeRefreshToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, revokeRefreshToken, token)
	return err
}
