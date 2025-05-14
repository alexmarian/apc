-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE password_reset_tokens (
                                       token TEXT PRIMARY KEY,
                                       login TEXT NOT NULL REFERENCES users(login),
                                       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                       expires_at TIMESTAMP NOT NULL,
                                       used_at TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE password_reset_tokens;
-- +goose StatementEnd