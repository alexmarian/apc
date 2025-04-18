-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE owners
(
    id                    INTEGER PRIMARY KEY,
    name                  TEXT    NOT NULL,
    normalized_name       TEXT    NOT NULL,
    identification_number TEXT    NOT NULL DEFAULT 'NAN', -- IDNP, fiscal code, etc.
    contact_phone         TEXT    NOT NULL DEFAULT 'NAN',
    contact_email         TEXT    NOT NULL DEFAULT 'NAN',
    first_detected_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    association_id        INTEGER NOT NULL REFERENCES associations (id) DEFAULT 1,
    created_at            TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
--     UNIQUE (identification_number)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE owners;
-- +goose StatementEnd
