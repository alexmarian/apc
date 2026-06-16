# Issue 001 — Move total owners stat inline [COMPLETED]

**Group:** A (OwnersReport UX)
**Type:** AFK
**Blocked by:** None

## What to build

The "Total Owners: X" stat box in `OwnersReport.vue` currently occupies its own row above the data table inside a `.summary-stats` div. Move it inline into the filter/controls row as plain text (e.g. `42 owners`) alongside the search input and sort dropdowns. Remove the `.summary-stats` div and its styles.

## Acceptance criteria

- [ ] Total owner count is displayed inline in the controls/filter bar, not in a separate row above the table
- [ ] The count reflects the current filtered result (same as before — `filteredSortedData.length`)
- [ ] The `.summary-stats` block and its associated CSS are removed
- [ ] No vertical space is consumed by the stat when the report loads
