---
status: completed
completed_at: 2026-06-02T00:00:00Z
implemented_by: claude
---

## What to build

A `VotingResultsWidget` Vue component in `@apc/voting-widgets` that displays voting results to a member after the gathering is tallied. Shows aggregate totals per matter alongside the member's own recorded choices, so they can verify their ballot was counted.

## Acceptance criteria

- [x] `VotingResultsWidget` exported from `@apc/voting-widgets` with props: `token: string`, `apiBaseUrl: string`
- [x] On mount, calls `GET /api/v1/member/gatherings/:token`
- [x] When `results` is null (gathering not yet `tallied`): displays a clear "results not yet available" state with the gathering status
- [x] When `results` is present: renders per-matter breakdown showing vote counts/percentages per option and pass/fail outcome
- [x] Highlights the member's own recorded choice on each matter (cross-referenced from `ballot` in the response)
- [x] If member did not vote, shows aggregate results without any personal highlight
- [x] Uses Naive UI component library

## Blocked by

- #003 — API: member read endpoints
- #005 — Shared package scaffold
