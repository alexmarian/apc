-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE apcs
(
    id               SERIAL PRIMARY KEY,
    cadastral_number TEXT UNIQUE,
    address          TEXT,
    total_area       NUMERIC,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE apcs;
-- +goose StatementEnd
