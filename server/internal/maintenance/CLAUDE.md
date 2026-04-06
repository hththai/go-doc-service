### Work on this maintenance folder to fix issues below:

### Issues

1. ScanFileFolder doesn't belong in this repository repository.go:223-265 — ScanFileFolder / scanFileFolder is file system logic with no DB interaction. A repository should abstract data access (DB), not perform file operations. This belongs in a service or a dedicated file utility layer.

2. getConfigDirectory is a no-op placeholder
repository.go:311-315 — It just returns the input unchanged. The comment says "replace with actual logic" but it's been left in, which is misleading.

3. buildFileMap skips hidden files inconsistently
GetTotalFileInStorage filters out .DS_Store via strings.HasPrefix(d.Name(), "."), but buildFileMap does not. Duplicate detection could include hidden files as false positives.

4. copyFile doesn't verify write success
repository.go:318-351 — out.Close() errors are silently swallowed via defer. On some filesystems, a flush error only surfaces on Close(), so you could silently copy a corrupt file. Use explicit close with error check.

5. GetFlagDocuments has a TODO
repository.go:140 — // TODO: review if need to keep this function. Stale TODOs in production code indicate incomplete review. Decide and either keep or remove it.

6. errMess pattern is non-idiomatic
Using a local errMess string variable as a format template is unusual and slightly harder to read than just writing fmt.Errorf("GetFlagDocuments: %w", err) inline.

### Note
should check unit test and integration test as well