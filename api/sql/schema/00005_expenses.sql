-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE expenses
(
    id          INTEGER PRIMARY KEY,
    amount      NUMERIC   NOT NULL,
    description TEXT      NOT NULL,
    destination TEXT      NOT NULL,
    date        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    month       INTEGER   NOT NULL,
    year        INTEGER   NOT NULL,
    category_id INTEGER   NOT NULL REFERENCES categories (id),
    account_id  INTEGER   NOT NULL REFERENCES accounts (id),
    created_at  TIMESTAMP          DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP          DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE expenses;
-- +goose StatementEnd
