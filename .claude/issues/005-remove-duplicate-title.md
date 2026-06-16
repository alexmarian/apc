# Issue 005 — Remove duplicate "Owners Report" title [COMPLETED]

**Group:** A (OwnersReport UX)
**Type:** AFK
**Blocked by:** None

## What to build

The owners report page (`pages/owners/report.vue`) renders an `NPageHeader` with the title "Owners Report". The `OwnersReport.vue` component then renders its own `NCard` also titled "Owners Report". This duplicates the heading at the top of the page.

Remove the `:title` prop from the `NCard` in `OwnersReport.vue`. The page header in `report.vue` is the single source of the page title.

## Acceptance criteria

- [ ] "Owners Report" appears exactly once on the page (in the `NPageHeader`)
- [ ] The `NCard` in `OwnersReport.vue` has no title
- [ ] The Export to CSV button (currently in the card `#header-extra` slot) is moved to the controls row or another appropriate location since the card header will be empty
