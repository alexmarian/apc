-- +goose Up
-- +goose StatementBegin
SELECT 'Adding voting_mode to gatherings table';

-- Add voting_mode column with default 'by_weight' for backward compatibility
ALTER TABLE gatherings
ADD COLUMN voting_mode TEXT NOT NULL DEFAULT 'by_weight'
CHECK (voting_mode IN ('by_weight', 'by_unit'));

CREATE INDEX idx_gatherings_voting_mode ON gatherings(voting_mode);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing voting_mode from gatherings table';

DROP INDEX IF EXISTS idx_gatherings_voting_mode;
ALTER TABLE gatherings DROP COLUMN voting_mode;

-- +goose StatementEnd
