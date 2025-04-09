-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE expenses
(
    id             INTEGER PRIMARY KEY,
    amount         NUMERIC,
    description    TEXT,
    destination    TEXT,
    date           TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    month          INTEGER,
    year           INTEGER,
    category_id    INTEGER REFERENCES categories (id),
    account_id     INTEGER REFERENCES accounts (id),
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE expenses;
-- +goose StatementEnd
