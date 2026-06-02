## What to build

Scaffold the `frontend/packages/voting-widgets` shared package so both the admin and member apps can import voting UI components from a single source. No widget logic yet — just the package structure, build config, and wiring into the monorepo.

## Acceptance criteria

- [x] `frontend/packages/voting-widgets/` created with `package.json` (name: `@apc/voting-widgets`), `tsconfig.json`, and Vite lib-mode `vite.config.ts`
- [x] Barrel export at `src/index.ts` (empty stubs for now)
- [x] Added to `frontend/package.json` workspaces and resolvable from admin and member apps via `"@apc/voting-widgets": "*"` (root already covers `packages/*`)
- [x] Both `frontend/apps/admin` and `frontend/apps/member` have the dependency declared in their `package.json`
- [x] `npm run build` in the package produces a valid dist; no build errors in either app

## Blocked by

None — can start immediately.

## Implementation

Created:
- `frontend/packages/voting-widgets/package.json` — name `@apc/voting-widgets`, lib exports to `dist/`
- `frontend/packages/voting-widgets/tsconfig.json` — extends `@vue/tsconfig/tsconfig.dom.json`
- `frontend/packages/voting-widgets/vite.config.ts` — Vite lib mode, externals vue
- `frontend/packages/voting-widgets/src/index.ts` — empty barrel export

Modified:
- `frontend/apps/admin/package.json` — added `"@apc/voting-widgets": "*"`
- `frontend/apps/member/package.json` — added `"@apc/voting-widgets": "*"`
