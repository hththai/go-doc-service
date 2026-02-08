# Client-Vite TODO

## Critical

- [ ] Fix broken API URL in `src/api/auth.tsx:23` - remove spaces from URL path
  ```typescript
  // Current (broken):
  `${import.meta.env.VITE_API_URL} / v1 / auth / refresh`
  // Should be:
  `${import.meta.env.VITE_API_URL}/v1/auth/refresh`
  ```

## Code Quality

- [ ] Extract `"tokenId"` localStorage key to a constant (used in `auth.tsx` and `upload.tsx`)
- [ ] Remove commented out code in `src/main.tsx` (lines 23-34)
- [ ] Clean up unused imports in `src/routes/signin.tsx`
- [ ] Address TODO comments:
  - `src/auth.tsx:1` - "store in api folder"
  - `src/api/upload.tsx:7` - "move this key to env"
- [ ] Fix unused description field in upload form - either remove it or include in FormData payload

## Missing Features

- [ ] Add form validation using TanStack Form (already installed)
- [ ] Implement error boundaries for graceful error handling
- [ ] Add user-facing error feedback for API failures
- [ ] Write tests with Vitest (already configured)

## Performance

- [ ] Add cleanup for `URL.createObjectURL()` in image preview to prevent memory leaks
- [ ] Implement request cancellation for pending operations

## Authentication

- [ ] Review dual token mechanism (tokenId in body AND localStorage) - consider simplifying
