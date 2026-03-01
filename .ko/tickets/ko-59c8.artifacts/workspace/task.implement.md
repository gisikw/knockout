## Summary

### Task 1 — `cmd_project.go:110`

Added a pre-eviction loop immediately before `reg.Projects[tag] = root`. For each existing entry in `reg.Projects`, if the path matches `root` but the key differs from `tag`, the stale entry is deleted from both `reg.Projects` and `reg.Prefixes`. If the evicted tag was `reg.Default`, the default is transferred to the incoming `tag`.

### Task 2 — `cmd_project_test.go`

Added `TestCmdProjectSetRetagEvictsOldTag` with two sub-cases:
- **retag removes old entry**: registers `projectDir` under `#foo`, then re-registers under `#bar`; asserts `len(reg.Projects) == 1`, `"bar"` exists with the correct path, and `"foo"` is absent.
- **retag transfers default**: registers under `#foo` with `--default`, then re-registers under `#bar`; asserts `reg.Default == "bar"`.

### Deviations
None. The `TestLoop` failures seen in the full test run are a pre-existing flaky test issue (timing-dependent, passes when run in isolation) and predate this change.
