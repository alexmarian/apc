-- +goose Up
ALTER TABLE voting_matters ADD COLUMN is_informative INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE voting_matters DROP COLUMN is_informative;
