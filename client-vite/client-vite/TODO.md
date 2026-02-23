# Client-Vite TODO

## Critical

- [x] Fix broken API URL in `src/api/auth.tsx:23` - remove spaces from URL path

## Code Quality

- [x] Extract `"tokenId"` localStorage key to a constant — `tokenId`/localStorage removed entirely, auth is now cookie-only
- [x] Remove commented out code in `src/main.tsx` (lines 48, 51, 55 — TanStackQueryProvider and RouterProvider remnants)
- [x] Clean up unused imports in `src/routes/signin.tsx` — all current imports are in use
- [x] Address `src/api/upload.tsx:7` TODO "move this key to env" — key no longer exists
- [ ] Address `src/auth.tsx:1` TODO "store in api folder" — file still lives at `src/auth.tsx`, not `src/api/`
- [x] Fix unused description field in upload form — description is included in FormData via the dynamic metadata loop

## Missing Features

- [ ] Add form validation using TanStack Form (already installed)
- [ ] Implement error boundaries for graceful error handling
- [x] Add user-facing error feedback for API failures (signin catches error silently with `console.error`)
- [x] Write tests with Vitest — tests exist: `Login.test.tsx`, `UploadFile.test.tsx`, `PurchaseInfo.test.tsx`, `api/auth.test.ts`

## Performance

- [x] Add cleanup for `URL.createObjectURL()` in image preview to prevent memory leaks (`UploadFile.tsx` lines 184, 362, 370 — no `revokeObjectURL` calls)
- [ ] Implement request cancellation for pending operations

## Authentication

- [x] Review dual token mechanism — resolved, tokenId/localStorage removed, fully cookie-based now
