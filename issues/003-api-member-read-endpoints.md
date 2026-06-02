## What to build

A new token-scoped read endpoint under `/api/v1/member/gatherings/:token` that resolves the opaque token to a member context and returns everything the member app needs to render: gathering metadata, the owner's eligible units, their submitted ballot (if any), and voting results (if the gathering is `tallied`).

No session, no JWT — the token in the path is the sole credential.

## Acceptance criteria

- [x] `GET /api/v1/member/gatherings/:token` resolves token via SHA-256 hash lookup against `member_invitations`
- [x] Returns 401 if token not found, revoked, or expired
- [x] Response includes: `gathering` (id, title, date, status, voting_mode), `owner` (name, identification), `units` (list of eligible unit slots for this owner in this gathering), `ballot` (null if not yet submitted, else the submitted `ballot_content` in read-only form), `results` (null unless gathering status is `tallied`, else the `voting_results` data)
- [x] Eligible units are resolved from `unit_slots` joined to `active_owner_units` view — same logic as existing ballot handler
- [x] No association ID or gathering ID in the URL — all context derived from token
- [x] New middleware function `MemberTokenAuth` that validates token and injects resolved context into request; reusable for issue #004

## Blocked by

- #001 — DB: `member_invitations` table
