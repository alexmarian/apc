-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE units
(
    id               INTEGER PRIMARY KEY,
    cadastral_number TEXT    NOT NULL UNIQUE,
    building_id      INTEGER    NOT NULL REFERENCES buildings (id),
    unit_number      TEXT    NOT NULL,
    address          TEXT    NOT NULL,
    entrance         INTEGER NOT NULL DEFAULT 1,
    area             NUMERIC NOT NULL,
    part             NUMERIC NOT NULL,
    unit_type        TEXT    NOT NULL DEFAULT 'apartment', -- apartment, commercial, etc.
    floor            INTEGER NOT NULL,
    room_count       INTEGER NOT NULL DEFAULT -1,
    created_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE units;
-- +goose StatementEnd
