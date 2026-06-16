# Issue 003 — Show selected owner name prominently in detail section [COMPLETED]

**Group:** A (OwnersReport UX)
**Type:** AFK
**Blocked by:** None

## What to build

When "Details" is clicked for an owner in `OwnersReport.vue`, the expanded detail section appears below the table but gives no clear indication of which owner is selected. The owner's name only appears buried in the "Owner's Units" sub-heading.

Add a prominent owner name header at the top of the detail section (above the type breakdown cards / units table), and visually highlight the selected row in the main table.

## Acceptance criteria

- [ ] A clear heading showing the selected owner's name appears at the top of the detail section, before any sub-tables
- [ ] The selected owner's row in the main table is visually distinguished (e.g. highlighted background or bold)
- [ ] The filter banner text ("Viewing details for selected owner") can be removed or replaced by the owner name heading — no duplicate labeling
