-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
alter table expenses add column document_ref varchar(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
alter table expenses drop column document_ref;
-- +goose StatementEnd
