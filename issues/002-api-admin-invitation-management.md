## What to build

Admin-facing API endpoints for managing member invitations for a gathering. The admin generates tokens, retrieves the list of existing invitations, and can revoke individual ones.

The plaintext token is returned **once** at creation time (the API stores only the hash). The admin app will display it for copy/export — it cannot be recovered after that response.

## Acceptance criteria

- [x] `POST /api/v1/associations/:association_id/gatherings/:gathering_id/invitations` — generates a cryptographically random opaque token, stores its SHA-256 hash in `member_invitations`, returns `{ token, expires_at, owner_id, invitation_id }`
- [x] Request body accepts `owner_id` and optional `expires_at` (defaults to 1 year from gathering date)
- [x] `GET /api/v1/associations/:association_id/gatherings/:gathering_id/invitations` — lists all invitations for the gathering with status (active / expired / revoked), never returns token plaintext
- [x] `DELETE /api/v1/associations/:association_id/gatherings/:gathering_id/invitations/:invitation_id` — sets `revoked_at`, token is immediately invalid
- [x] All endpoints require existing admin JWT auth middleware
- [x] Returns 409 if an active invitation already exists for that owner + gathering

## Blocked by

- #001 — DB: `member_invitations` table
