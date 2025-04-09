-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE units
(
    id               INTEGER PRIMARY KEY,
    cadastral_number TEXT UNIQUE,
    building_id      TEXT    NOT NULL REFERENCES buildings (id),
    unit_number      TEXT,
    address          TEXT,
    area             NUMERIC NOT NULL,
    part             NUMERIC NOT NULL,
    unit_type        TEXT, -- apartment, commercial, etc.
    floor            INTEGER,
    room_count       INTEGER,
    created_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE units;
-- +goose StatementEnd
