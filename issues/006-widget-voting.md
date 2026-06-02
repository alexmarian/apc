## What to build

A `VotingWidget` Vue component in `@apc/voting-widgets` that handles the full voting flow for a member: renders an interactive ballot form when the member hasn't voted yet, and switches to a read-only receipt showing their recorded choices after submission.

This replaces the admin-app-only `VotingInterface.vue` for the member context, consuming the token-scoped API from #003 and #004.

## Acceptance criteria

- [x] `VotingWidget` exported from `@apc/voting-widgets` with props: `token: string`, `apiBaseUrl: string`
- [x] On mount, calls `GET /api/v1/member/gatherings/:token`; displays owner name, unit count, and voting weight in a summary header
- [x] When `ballot` is null in the response: renders interactive ballot form supporting `yes_no`, `single_choice`, and `multiple_choice` matter types (matching existing `VotingInterface.vue` logic)
- [x] Submit calls `POST /api/v1/member/gatherings/:token/ballot`; on success transitions to read-only receipt without page reload
- [x] When `ballot` is already present in the response: renders read-only receipt showing submitted choices per matter (no form controls)
- [x] Displays appropriate message when gathering is not `active` (not yet open, or already closed)
- [x] Uses Naive UI component library (consistent with admin app)

## Blocked by

- #003 — API: member read endpoints
- #004 — API: member ballot submission
- #005 — Shared package scaffold

## Implementation

Backend change:
- `api/internal/handlers/member_handler.go` — added `memberMatterInfo` struct and `Matters []memberMatterInfo` field to `memberGatheringResponse`; `HandleGetMemberContext` now fetches voting matters via `GetVotingMatters` and includes them in the response

Frontend:
- `frontend/packages/voting-widgets/src/VotingWidget.vue` — full component: summary header, inactive-gathering message, interactive ballot form (yes_no/single_choice/multiple_choice), post-submit read-only receipt
- `frontend/packages/voting-widgets/src/index.ts` — exports `VotingWidget`
- `frontend/packages/voting-widgets/vite.config.ts` — `naive-ui` added to externals
- `frontend/packages/voting-widgets/package.json` — `naive-ui` added as peer + dev dep
- `frontend/apps/member/package.json` — `naive-ui` added as dependency
