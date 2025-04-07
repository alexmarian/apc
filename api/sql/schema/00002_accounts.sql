-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE accounts
(
    id          SERIAL PRIMARY KEY,
    destination TEXT,
    description TEXT,
    apc_id      INTEGER REFERENCES apc (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE accounts;
-- +goose StatementEnd
