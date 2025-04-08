-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE buildings
(
    id             SERIAL PRIMARY KEY,
    name           TEXT,
    address        TEXT,
    total_area     NUMERIC,
    association_id INTEGER REFERENCES associations (id),
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE buildings;
-- +goose StatementEnd