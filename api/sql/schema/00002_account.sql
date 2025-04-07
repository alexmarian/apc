-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE account
(
    id          SERIAL PRIMARY KEY,
    destination TEXT UNIQUE,
    apc_id      INTEGER REFERENCES apc (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE account;
-- +goose StatementEnd
