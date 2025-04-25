-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE accounts
(
    id             INTEGER PRIMARY KEY,
    number         TEXT    NOT NULL,
    destination    TEXT    NOT NULL,
    description    TEXT    NOT NULL,
    is_active      BOOL    NOT NULL default true,
    association_id INTEGER NOT NULL REFERENCES associations (id),
    created_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP        DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE accounts;
-- +goose StatementEnd
