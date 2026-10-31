# Clean up maintenance repository error handling

## Goal
Resolve the open review issues listed in `server/internal/maintenance/CLAUDE.md`.

## Issues
1. `GetFlagDocuments` in `server/internal/maintenance/repository.go` has a stale
   `// TODO: review if need to keep this function`. Find every caller (search the
   whole server/ folder). If it is used, keep it and remove the TODO; if it is
   unused, remove the function and its tests.
2. The local `errMess` string used as a format template is non-idiomatic. Replace
   it with inline wrapping, e.g. `fmt.Errorf("GetFlagDocuments: %w", err)`, across
   the maintenance package. Error messages must keep the same meaning.

## Constraints
- Behaviour must not change.
- Keep unit tests (`internal/maintenance/test`) and integration tests
  (`internal/maintenance/test/integration`) passing; update them if messages change.
- When finished, tick these issues off in `server/internal/maintenance/CLAUDE.md`.

## Done when
- No `errMess` remains in `server/internal/maintenance`.
- The TODO is resolved.
- `go test ./...` passes in server/.
