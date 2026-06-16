# Issue 002 — Filtered columns reflect filtered units [COMPLETED]

**Group:** A (OwnersReport UX)
**Type:** AFK
**Blocked by:** None

## What to build

When a unit-type filter is active in `OwnersReport.vue`, the Area, Part, and Units columns in the main owners table show the owner's total across all unit types (from `item.statistics.*`). They should instead compute from only the units matching the active filter.

Unit data is already loaded whenever `unitTypeFilter` is non-empty (the `unitsNeeded` computed forces it). No API changes are needed — this is a purely client-side computed column change.

## Acceptance criteria

- [ ] When no unit-type filter is active, columns show full totals (existing behaviour)
- [ ] When a unit-type filter is active, Area shows sum of `unit.area` for matching units only
- [ ] When a unit-type filter is active, Part shows sum of `unit.part` for matching units only
- [ ] When a unit-type filter is active, Units count shows count of matching units only
- [ ] Clearing the filter restores full totals immediately (reactive)
- [ ] CSV export reflects the same filtered values when a filter is active
