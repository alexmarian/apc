-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE expenses
(
    id          SERIAL PRIMARY KEY,
    amount      NUMERIC,
    description TEXT,
    date        TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    apc_id      INTEGER REFERENCES apc (id),
    category_id INTEGER REFERENCES categories (id),
    account_id  INTEGER REFERENCES accounts (id),
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE expenses;
-- +goose StatementEnd
