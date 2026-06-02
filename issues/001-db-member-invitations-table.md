## What to build

Add a `member_invitations` table to the schema that backs the opaque per-member-per-gathering tokens. This is the foundation for all member app access control.

The table stores a hash of the token (never the plaintext), the owner and gathering it grants access to, an admin-configured expiry date, and a soft-delete revocation flag that integrates with the existing `revoked_tokens` pattern.

## Acceptance criteria

- [x] Goose migration creates `member_invitations` table with columns: `id`, `gathering_id` (FK → `gatherings`), `owner_id` (FK → `owners`), `token_hash` (TEXT, unique), `expires_at` (TIMESTAMP), `revoked_at` (TIMESTAMP NULL), `created_at`, `updated_at`
- [x] Unique index on `(gathering_id, owner_id)` — one active invitation per owner per gathering
- [x] Index on `token_hash` for fast token lookup
- [x] sqlc queries: `CreateMemberInvitation`, `GetMemberInvitationByTokenHash`, `ListMemberInvitationsByGathering`, `RevokeMemberInvitation`
- [x] Migration rolls back cleanly

## Blocked by

None — can start immediately.
