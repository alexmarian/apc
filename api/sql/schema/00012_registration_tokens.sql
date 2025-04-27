-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE registration_tokens
(
    token       TEXT PRIMARY KEY,
    created_by  TEXT NOT NULL REFERENCES users (login),
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at  TIMESTAMP NOT NULL,
    used_at     TIMESTAMP,
    used_by     TEXT REFERENCES users (login),
    revoked_at  TIMESTAMP,
    revoked_by  TEXT REFERENCES users (login),
    description TEXT NOT NULL,
    is_admin    BOOLEAN NOT NULL DEFAULT FALSE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE registration_tokens;
-- +goose StatementEnd