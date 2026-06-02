-- +goose Up
ALTER TABLE voting_matters ADD COLUMN title_ru TEXT NOT NULL DEFAULT '';
ALTER TABLE voting_matters ADD COLUMN description_ru TEXT;

-- +goose Down
ALTER TABLE voting_matters DROP COLUMN title_ru;
ALTER TABLE voting_matters DROP COLUMN description_ru;
