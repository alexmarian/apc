---
status: completed
completed_at: 2026-06-02T00:00:00Z
implemented_by: claude
---

## What to build

An invitation management panel in the admin app for a gathering, allowing admins to generate per-owner member links, copy or export them for distribution, set expiry dates, and revoke individual invitations.

This is the admin's primary tool for getting voting links into members' hands (email sending is deferred).

## Acceptance criteria

- [x] New `MemberInvitationsPanel` component added to the gathering detail view in the admin app
- [x] Lists all owners qualified for the gathering (from existing `VotingOwnersReport` / eligible voters logic) with their invitation status: not generated / active / expired / revoked
- [x] "Generate link" button per owner: calls `POST .../invitations`, displays the returned plaintext token as a copyable URL (`https://member.blocul-nostru.online/:token`); warns that the link cannot be recovered after closing the dialog
- [x] "Generate all" bulk action: generates links for all owners without an active invitation in one flow
- [x] Copy-to-clipboard button per link
- [x] "Export CSV" exports all active links (owner name, link, expiry) for the gathering
- [x] "Revoke" button per invitation: calls `DELETE .../invitations/:id`, updates status in the list immediately
- [x] Expiry date field on generation (defaults to 1 year from gathering date, editable)

## Blocked by

- #002 — API: admin invitation management endpoints
