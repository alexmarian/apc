# Issue 004 — Unit type breakdown as horizontal scrollable cards [COMPLETED]

**Group:** A (OwnersReport UX)
**Type:** AFK
**Blocked by:** None

## What to build

Replace the `NDataTable` per-type summary in `OwnersReport.vue` (columns: Type / Count / Area / Part) with a row of compact cards — one per unit type the selected owner holds. The row is horizontally scrollable and does not wrap. Each card shows: unit type name, count, area (m²), and part (%). Clicking a card filters the units table below to that type (same behaviour as clicking a row currently). The Total summary row disappears — it is already visible in the main owners table.

## Acceptance criteria

- [ ] Per-type summary is rendered as a horizontal scrollable row of cards, not a table
- [ ] Each card shows: type name, unit count, total area, total part
- [ ] Cards only appear for unit types the selected owner actually holds (no empty cards)
- [ ] Clicking a card filters the units detail table to that type (same toggle behaviour as before)
- [ ] The active/selected card is visually highlighted
- [ ] The card row scrolls horizontally on overflow, does not wrap to a second line
- [ ] The old `NDataTable` summary and its Total row are removed
