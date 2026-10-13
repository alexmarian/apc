-- +goose Up
-- +goose StatementBegin
-- Which buildings a gathering applies to (JSON array of building ids, like
-- the other qualification lists). NULL: every building of the association.
-- With several buildings it is still one decision: their qualifying units
-- vote together.
ALTER TABLE gatherings ADD COLUMN qualification_buildings TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE gatherings DROP COLUMN qualification_buildings;
-- +goose StatementEnd
