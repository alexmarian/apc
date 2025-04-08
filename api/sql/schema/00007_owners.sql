-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE owners
(
    id                    SERIAL PRIMARY KEY,
    name                  TEXT NOT NULL,
    normalized_name       TEXT NOT NULL,
    identification_number TEXT, -- IDNP, fiscal code, etc.
    contact_phone         TEXT,
    contact_email         TEXT,
    first_detected_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (normalized_name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE owners;
-- +goose StatementEnd
