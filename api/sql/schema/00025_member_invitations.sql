-- +goose Up
-- +goose StatementBegin
CREATE TABLE member_invitations (
    id           INTEGER PRIMARY KEY,
    gathering_id INTEGER  NOT NULL REFERENCES gatherings (id) ON DELETE CASCADE,
    owner_id     INTEGER  NOT NULL REFERENCES owners (id) ON DELETE CASCADE,
    token_hash   TEXT     NOT NULL UNIQUE,
    expires_at   DATETIME NOT NULL,
    revoked_at   DATETIME NULL,
    created_at   DATETIME NOT NULL DEFAULT (datetime('now')),
    updated_at   DATETIME NOT NULL DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_member_invitations_owner_gathering
    ON member_invitations (gathering_id, owner_id)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_member_invitations_token_hash ON member_invitations (token_hash);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_member_invitations_token_hash;
DROP INDEX IF EXISTS idx_member_invitations_owner_gathering;
DROP TABLE IF EXISTS member_invitations;
-- +goose StatementEnd
