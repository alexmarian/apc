-- +goose Up
-- +goose StatementBegin
-- Owners as app users: an account type (staff keep mandatory TOTP; owner
-- accounts may sign in with a password only until they enable TOTP), a
-- profile photo, the user <-> owner links (one account per owner record; an
-- account may hold owner records of several associations) and one-time
-- activation links an admin sends to an owner.
ALTER TABLE users ADD COLUMN account_type TEXT NOT NULL DEFAULT 'staff' CHECK (account_type IN ('staff', 'owner'));
ALTER TABLE users ADD COLUMN photo_path TEXT;
ALTER TABLE users ADD COLUMN photo_updated_at TIMESTAMP;

CREATE TABLE user_owners
(
    user_id    INTEGER   NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    owner_id   INTEGER   NOT NULL UNIQUE REFERENCES owners (id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, owner_id)
);

CREATE TABLE owner_activations
(
    id         INTEGER PRIMARY KEY,
    owner_id   INTEGER   NOT NULL REFERENCES owners (id) ON DELETE CASCADE,
    token_hash TEXT      NOT NULL UNIQUE,
    expires_at TIMESTAMP NOT NULL,
    created_by TEXT      NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at    TIMESTAMP,
    revoked_at TIMESTAMP
);
CREATE INDEX idx_owner_activations_owner ON owner_activations (owner_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE owner_activations;
DROP TABLE user_owners;
ALTER TABLE users DROP COLUMN photo_updated_at;
ALTER TABLE users DROP COLUMN photo_path;
ALTER TABLE users DROP COLUMN account_type;
-- +goose StatementEnd
