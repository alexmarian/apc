-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE units
(
    cadastral_number          VARCHAR(50) PRIMARY KEY,
    building_cadastral_number VARCHAR(50)    NOT NULL REFERENCES buildings (cadastral_number),
    unit_number               VARCHAR(50),
    address                   TEXT,
    area                      NUMERIC NOT NULL,
    unit_type                 TEXT, -- apartment, commercial, etc.
    floor                     INTEGER,
    room_count                INTEGER,
    created_at                TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at                TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE units;
-- +goose StatementEnd
