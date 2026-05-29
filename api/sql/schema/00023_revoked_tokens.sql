-- +goose Up
-- +goose StatementBegin
CREATE TABLE revoked_tokens (
    jti        TEXT PRIMARY KEY,
    revoked_at DATETIME NOT NULL DEFAULT (datetime('now')),
    expires_at DATETIME NOT NULL
);
CREATE INDEX idx_revoked_tokens_expires_at ON revoked_tokens(expires_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE revoked_tokens;
-- +goose StatementEnd
