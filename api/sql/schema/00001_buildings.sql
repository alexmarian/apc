-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE buildings
(
    id               INTEGER PRIMARY KEY,
    name             TEXT    NOT NULL,
    address          TEXT    NOT NULL,
    cadastral_number TEXT    NOT NULL,
    total_area       NUMERIC NOT NULL,
    association_id   INTEGER NOT NULL REFERENCES associations (id),
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE buildings;
-- +goose StatementEnd