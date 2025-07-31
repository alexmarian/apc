-- +goose Up
-- +goose StatementBegin
SELECT 'Adding location field to gatherings table';

-- Add location column to gatherings table
ALTER TABLE gatherings ADD COLUMN location TEXT NOT NULL DEFAULT '';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'Removing location field from gatherings table';

-- Remove location column from gatherings table
ALTER TABLE gatherings DROP COLUMN location;

-- +goose StatementEnd