-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories ADD COLUMN original_labels TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories DROP COLUMN original_labels;
-- +goose StatementEnd
