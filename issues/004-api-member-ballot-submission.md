## What to build

The member-facing ballot submission endpoint. A member submits their votes via their token link. On success, a `GatheringParticipant` record is created, unit slots are assigned, a ballot is stored, and the member is automatically checked in — no prior check-in step required.

## Acceptance criteria

- [x] `POST /api/v1/member/gatherings/:token/ballot` — accepts `{ ballot_content: { [matter_id]: BallotVote } }`
- [x] Uses `MemberTokenAuth` middleware from #003 to resolve token context
- [x] Returns 400 if gathering is not `active`
- [x] Returns 409 if a valid ballot already exists for this owner in this gathering (one ballot per owner, no resubmission)
- [x] Creates `GatheringParticipant` with `check_in_time` set to submission time (auto check-in)
- [x] Assigns unit slots to the new participant (same logic as existing `ballot_handler.go`)
- [x] Stores ballot with SHA-256 hash; triggers async tally update
- [x] Returns `{ ballot_id, ballot_hash, submitted_at }` — member app uses this to switch to read-only receipt state
- [x] Audit log entry created for the submission

## Blocked by

- #003 — API: member read endpoints (provides `MemberTokenAuth` middleware)

## Implementation

Implemented in:
- `api/internal/handlers/gathering/handlers/member_ballot_handler.go` — `MemberBallotHandler` with `HandleSubmitMemberBallot`
- `api/internal/handlers/gathering/router.go` — `MemberBallot` field added to `GatheringRouter`
- `api/internal/handlers/member_handler.go` — exported `MemberInvitationFromContext` helper
- `api/main.go` — route registered: `POST /v1/api/member/gatherings/{memberToken}/ballot`
