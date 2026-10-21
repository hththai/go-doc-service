### Work on this maintenance folder to fix issues below:

### Issues

1. GetFlagDocuments has a TODO
repository.go:140 — // TODO: review if need to keep this function. Stale TODOs in production code indicate incomplete review. Decide and either keep or remove it.

2. errMess pattern is non-idiomatic
Using a local errMess string variable as a format template is unusual and slightly harder to read than just writing fmt.Errorf("GetFlagDocuments: %w", err) inline.

### Note
should check unit test and integration test as well