# Issue 007 — Owner detail page `/owners/:id`

**Group:** B (Owner detail page + co-ownership)
**Type:** AFK
**Blocked by:** Issue 006

## What to build

Create a new page at `/owners/:id` that shows a single owner's full profile: their stats (total units, area, part), all units across all buildings, and co-owners per unit. The inline "Details" drill-down in `OwnersReport.vue` is replaced by navigation to this page. The "Details" button in the main owners table becomes a router-link to `/owners/:id`.

Browser back handles return navigation. The page header shows a breadcrumb derived from the `from` route (e.g. `Owners Report > Owner A`). The `is_billing` flag (from Issue 006) is shown on co-owned units to indicate which owner is the billing contact.

The page reuses the existing API (`GetOwnerUnitsWithDetailsForReport` query via the owner report endpoint filtered by owner ID, or a dedicated endpoint if needed).

## Acceptance criteria

- [x] Route `/owners/:id` renders an owner detail page
- [x] Page shows owner name, identification number, contact phone, contact email
- [x] Page shows summary stats: total units, total area, total condo part
- [x] Page lists all active units across all buildings in a sortable table
- [x] For each unit with co-owners, co-owner names are displayed
- [ ] `is_billing` status is shown per co-owner on shared units (blocked by Issue 006)
- [x] "Details" button in `OwnersReport.vue` navigates to `/owners/:id` instead of expanding inline
- [x] Inline drill-down section in `OwnersReport.vue` is removed
- [x] Page header breadcrumb reflects the navigation source (`from` query param preserved)
- [x] Browser back returns to the owners report at the same scroll/filter position
