-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE associations
(
    id            INTEGER PRIMARY KEY,
    name          TEXT NOT NULL ,
    address       TEXT NOT NULL ,
    administrator TEXT NOT NULL ,
    created_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE associations;
-- +goose StatementEnd
