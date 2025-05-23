// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: password_reset_tokens.sql

package database

import (
	"context"
	"time"
)

const createPasswordResetToken = `-- name: CreatePasswordResetToken :exec
INSERT INTO password_reset_tokens (token,
                                   login,
                                   created_at,
                                   expires_at)
VALUES (?, ?, datetime('now'), ?)
`

type CreatePasswordResetTokenParams struct {
	Token     string
	Login     string
	ExpiresAt time.Time
}

func (q *Queries) CreatePasswordResetToken(ctx context.Context, arg CreatePasswordResetTokenParams) error {
	_, err := q.db.ExecContext(ctx, createPasswordResetToken, arg.Token, arg.Login, arg.ExpiresAt)
	return err
}

const getValidPasswordResetToken = `-- name: GetValidPasswordResetToken :one
SELECT token, login, created_at, expires_at, used_at
FROM password_reset_tokens
WHERE token = ?
  AND used_at IS NULL
  AND expires_at > datetime('now') LIMIT 1
`

func (q *Queries) GetValidPasswordResetToken(ctx context.Context, token string) (PasswordResetToken, error) {
	row := q.db.QueryRowContext(ctx, getValidPasswordResetToken, token)
	var i PasswordResetToken
	err := row.Scan(
		&i.Token,
		&i.Login,
		&i.CreatedAt,
		&i.ExpiresAt,
		&i.UsedAt,
	)
	return i, err
}

const usePasswordResetToken = `-- name: UsePasswordResetToken :exec
UPDATE password_reset_tokens
SET used_at = datetime('now')
WHERE token = ?
`

func (q *Queries) UsePasswordResetToken(ctx context.Context, token string) error {
	_, err := q.db.ExecContext(ctx, usePasswordResetToken, token)
	return err
}
