# Reject refresh tokens used as access tokens

## Goal
Fix pentest finding vuln-0001 (HIGH): a refresh token is currently accepted
where an access token is required.

## Context
- Token creation and validation: `server/internal/auth/service.go`
  (`CreateTokensForUser`, `RefreshAccessToken`, `ValidateAccessToken`).
- Middleware: `server/middleware/authen/`.
- Existing tests: `server/internal/auth/test/auth_test.go`.

## Requirements
1. Access and refresh tokens must carry a distinguishing claim (for example
   `typ: "access"` / `typ: "refresh"`).
2. `ValidateAccessToken` (and the auth middleware) must reject a token whose
   type is not `access`. Refresh must reject a token whose type is not `refresh`.
3. First write a failing test that uses a refresh token as an access token,
   then fix the code so it passes.

## Constraints
- Keep the cookie-based flow and existing endpoints unchanged.
- Do not log or return token contents in errors.

## Done when
- The new test passes, all existing auth tests pass, `go test ./...` passes.
