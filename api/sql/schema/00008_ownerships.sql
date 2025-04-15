-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE ownerships
(
    id                    INTEGER PRIMARY KEY,
    unit_id               INTEGER   NOT NULL REFERENCES units (id),
    owner_id              INTEGER   NOT NULL REFERENCES owners (id),
    association_id        INTEGER   NOT NULL REFERENCES associations (id) DEFAULT 1,
    start_date            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    end_date              TIMESTAMP DEFAULT NULL, -- NULL means current ownership
    is_active             BOOLEAN   DEFAULT TRUE,
    registration_document TEXT      NOT NULL,
    registration_date     TIMESTAMP NOT NULL,
    created_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at            TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_active_ownership UNIQUE (unit_id, owner_id, is_active)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE ownerships;
-- +goose StatementEnd
