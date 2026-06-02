---
status: completed
completed_at: 2026-06-02T00:00:00Z
implemented_by: claude
---

## What to build

Wire up the member app (`frontend/apps/member`) to be a real single-page application that reads the opaque token from the URL and renders the appropriate widget. The token in the URL is the entire auth mechanism — no login screen, no session.

## Acceptance criteria

- [x] Member app has a single route that captures the token from the URL path (e.g. `/:token`)
- [x] On load, calls `GET /api/v1/member/gatherings/:token` to determine gathering state
- [x] Renders `VotingWidget` when gathering is `active` and member hasn't voted, or when they have (read-only receipt mode is handled inside the widget)
- [x] Renders `VotingResultsWidget` when gathering is `tallied`
- [x] Renders an informative state for `draft`/`published`/`closed` statuses ("Voting has not started yet", "Voting has ended, results pending")
- [x] Renders a clear error state for invalid/expired/revoked tokens (401 response)
- [x] App is production-built and served by the existing `member` Caddy config (`caddy/sites/member.conf`)

## Blocked by

- #006 — Voting widget
- #007 — Results widget
