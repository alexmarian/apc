-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
ALTER TABLE ownerships ADD COLUMN is_voting BOOLEAN NOT NULL DEFAULT false;
CREATE UNIQUE INDEX idx_voting_owner_per_unit
    ON ownerships (unit_id)
    WHERE is_voting = true;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
ALTER TABLE ownerships DROP COLUMN is_voting;
DROP INDEX idx_voting_owner_per_unit;
-- +goose StatementEnd
