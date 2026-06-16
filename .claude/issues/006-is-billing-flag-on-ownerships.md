# Issue 006 — Add `is_billing` flag to ownerships

**Group:** B (Owner detail page + co-ownership)
**Type:** AFK
**Blocked by:** None

## What to build

Add an `is_billing BOOLEAN NOT NULL DEFAULT false` column to the `ownerships` table, following the same pattern as the existing `is_voting` flag (schema `00013_ownership_voting.sql`). Add a unique partial index ensuring at most one billing owner per unit at a time.

Expose `is_billing` in the ownership API (read and write): the `CreateOwnership` and `UpdateOwnership` queries, the Go handler structs, and the frontend `OwnershipForm.vue` (a checkbox, same as is_voting).

## Acceptance criteria

- [ ] New goose migration adds `is_billing` column with default false
- [ ] Unique partial index prevents more than one `is_billing = true` per `unit_id`
- [ ] `CreateOwnership` and `DeactivateOwnership`/update queries include `is_billing`
- [ ] Go handler correctly reads and writes `is_billing`
- [ ] `OwnershipForm.vue` shows an "Is Billing Contact" checkbox (same pattern as voting)
- [ ] Existing ownerships default to `is_billing = false` (no data migration needed)
