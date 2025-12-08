-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query - adding category performance indexes';

-- Index for filtering active/inactive categories per association
CREATE INDEX IF NOT EXISTS idx_categories_association_deleted
ON categories(association_id, is_deleted);

-- Composite index for uniqueness checks and efficient lookups
CREATE INDEX IF NOT EXISTS idx_categories_lookup
ON categories(association_id, type, family, name)
WHERE is_deleted = FALSE;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query - removing category indexes';

DROP INDEX IF EXISTS idx_categories_association_deleted;
DROP INDEX IF EXISTS idx_categories_lookup;

-- +goose StatementEnd
