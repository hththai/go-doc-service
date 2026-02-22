# Auth Session Persistence Fix

## Problem

After logging in, closing and reopening a browser tab would redirect the user back to the login page — even though both cookies were still valid.

## Root Cause

The issue was a **race condition between async auth resolution and TanStack Router's route guard**.

### Flow before fix (broken)

1. Tab reopens → React Query cache is empty (in-memory only)
2. `useCurrentUser` starts fetching `GET /users/me` asynchronously
3. During the fetch: `isLoading = true`, `data = undefined` → `isAuthenticated = false`
4. `beforeLoad` in `_auth.tsx` sees `isAuthenticated = false` → **immediately redirects to `/signin`**
5. `getMe()` eventually resolves (with valid tokens), but the user is already on the login page

The cookies themselves were fine — both are **persistent cookies** (`MaxAge` is set server-side), so they survive tab closes. The problem was purely a frontend timing issue.

### Why the session should have worked

| Cookie          | MaxAge           | Survives tab close? |
| --------------- | ---------------- | ------------------- |
| `access_token`  | 600s (10 min)    | Yes                 |
| `refresh_token` | 604800s (1 week) | Yes                 |

`getMe()` in [src/api/auth.tsx](src/api/auth.tsx) already handles token refresh:

```
GET /users/me → 401 → POST /refresh → retry GET /users/me
```

So even an expired access token is handled transparently.

## Fix

Three changes across two files.

### 1. `_auth.tsx` — Skip redirect while auth is still loading

```tsx
// beforeLoad
beforeLoad: ({ context, location }) => {
  if (context.auth.isLoading) return; // wait for auth check to complete
  if (!context.auth.isAuthenticated) {
    throw redirect({ to: "/signin", search: { redirect: location.href } });
  }
},
```

### 2. `_auth.tsx` — Guard the layout component against rendering unauth'd content

```tsx
function AuthLayout() {
  ...
  if (auth.isLoading || !auth.isAuthenticated) return null;
  ...
}
```

### 3. `main.tsx` — Re-run `beforeLoad` after auth resolves

TanStack Router does **not** automatically re-run `beforeLoad` when context changes. Without this, skipping the redirect during loading would leave unauthenticated users permanently on the protected route.

```tsx
function InnerApp() {
  const auth = useAuth();

  useEffect(() => {
    router.invalidate(); // re-runs beforeLoad with the resolved auth state
  }, [auth.isLoading, auth.isAuthenticated]);

  return <RouterProvider router={router} context={{ auth }} />;
}
```

## Full Flow After Fix

### Returning user (valid cookies)

| Step                                     | `isLoading` | `isAuthenticated` | What happens                              |
| ---------------------------------------- | ----------- | ----------------- | ----------------------------------------- |
| Page loads                               | `true`      | `false`           | `beforeLoad` waits, layout renders `null` |
| `getMe()` resolves (+ refresh if needed) | `false`     | `true`            | `useEffect` fires → `router.invalidate()` |
| `beforeLoad` re-runs                     | `false`     | `true`            | Passes → protected content renders        |

### No session (logged out or expired refresh token)

| Step                     | `isLoading` | `isAuthenticated` | What happens                              |
| ------------------------ | ----------- | ----------------- | ----------------------------------------- |
| Page loads               | `true`      | `false`           | `beforeLoad` waits, layout renders `null` |
| `getMe()` returns `null` | `false`     | `false`           | `useEffect` fires → `router.invalidate()` |
| `beforeLoad` re-runs     | `false`     | `false`           | Redirects to `/signin`                    |

### Logout

Already handled — the `handleLogout` function in `AuthLayout` explicitly calls `router.invalidate()` after clearing cookies.
