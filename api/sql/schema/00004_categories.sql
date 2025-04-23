-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE categories
(
    id             INTEGER PRIMARY KEY,
    type           TEXT    NOT NULL,
    family         TEXT    NOT NULL,
    name           TEXT    NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    association_id INTEGER NOT NULL REFERENCES associations (id),
    created_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE categories;
-- +goose StatementEnd
