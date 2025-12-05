-- +goose Up
-- +goose StatementBegin
SELECT 'Removing unique voting owner constraint - allow all joint owners to vote';

-- Drop the unique index that restricts one voting owner per unit
DROP INDEX IF EXISTS idx_voting_owner_per_unit;

-- Set all active ownerships as voting-eligible
-- This makes is_voting redundant but we'll keep the column for backward compatibility
UPDATE ownerships
SET is_voting = TRUE
WHERE is_active = TRUE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Restoring unique voting owner constraint';

-- Reset is_voting to false for all
UPDATE ownerships
SET is_voting = FALSE;

-- Recreate the unique index
CREATE UNIQUE INDEX idx_voting_owner_per_unit
    ON ownerships (unit_id)
    WHERE is_voting = true;

-- +goose StatementEnd
