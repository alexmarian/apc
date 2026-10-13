-- +goose Up
-- TOTP onboarding: a user is "pending" until they confirm a code from their
-- authenticator. Existing users already log in with TOTP, so they count as
-- confirmed.
ALTER TABLE users ADD COLUMN totp_confirmed_at TIMESTAMP;
UPDATE users SET totp_confirmed_at = created_at;

-- Disabling blocks login, refresh and API access without deleting the user
-- (registration tokens, audit logs etc. keep referencing the login).
ALTER TABLE users ADD COLUMN disabled_at TIMESTAMP;
ALTER TABLE users ADD COLUMN disabled_by TEXT;

-- Associations an invited user is linked to when registering with a token.
CREATE TABLE registration_token_associations
(
    token          TEXT    NOT NULL REFERENCES registration_tokens (token),
    association_id INTEGER NOT NULL REFERENCES associations (id),
    PRIMARY KEY (token, association_id)
);

-- +goose Down
DROP TABLE registration_token_associations;
ALTER TABLE users DROP COLUMN disabled_by;
ALTER TABLE users DROP COLUMN disabled_at;
ALTER TABLE users DROP COLUMN totp_confirmed_at;
