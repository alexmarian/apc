# Issue 008 — Co-owner linking on owner detail page

**Group:** B (Owner detail page + co-ownership)
**Type:** AFK
**Blocked by:** Issue 007
**Status:** completed
**Completed at:** 2026-06-16

## What to build

On the `/owners/:id` page (built in Issue 007), co-owner names displayed on shared units should be clickable links to `/owners/:co-owner-id`. This enables natural chain navigation between co-owners using browser back, without any special back-link logic.

The breadcrumb should update to reflect the chain (e.g. `Owners Report > Owner A > Owner B`) as the user follows co-owner links.

## Acceptance criteria

- [x] Each co-owner name on the owner detail page is a router-link to `/owners/:co-owner-id`
- [x] Navigating to a co-owner's page shows their full profile (all their units, their co-owners, etc.)
- [x] The co-owner's page also has co-owner links (fully recursive)
- [x] Browser back at each step returns to the previous owner's page
- [x] Breadcrumb reflects the navigation chain up to a reasonable depth (e.g. 3 levels)

## Implementation

Changes made to `frontend/apps/admin/src/pages/owners/[ownerId]/index.vue`:

- `ownerId` converted to a `computed` so the `watch` fires on param changes (enabling same-component re-render when navigating between co-owner pages without a full remount)
- Added `breadcrumbChain` computed that parses a `chain` query param (JSON array of `{id, name}` entries)
- Co-owner name column now renders a styled `<a>` link; clicking it pushes to `/owners/:coOwnerId` with the updated `chain` param (capped at 3 ancestors)
- `NPageHeader` header slot renders an `NBreadcrumb`: `Owners Report > [chain entries] > current owner`; each ancestor is clickable and navigates back correctly
- `handleBack` and `navigateToBreadcrumb` reconstruct the chain when jumping back through the breadcrumb
- i18n: added `owners.detail.backToPrevOwner` key to en/ro/ru locale files
