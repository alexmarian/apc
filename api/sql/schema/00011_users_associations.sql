-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
CREATE TABLE users_associations
(
    id             INTEGER PRIMARY KEY,
    user_id        INTEGER   NOT NULL,
    association_id INTEGER   NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users (id),
    FOREIGN KEY (association_id) REFERENCES associations (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
DROP TABLE users_associations;
-- +goose StatementEnd