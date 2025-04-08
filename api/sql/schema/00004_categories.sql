-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE categories
(
    id             SERIAL PRIMARY KEY,
    type           TEXT,
    family         TEXT,
    name           TEXT,
    association_id INTEGER REFERENCES associations (id),
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE categories;
-- +goose StatementEnd
