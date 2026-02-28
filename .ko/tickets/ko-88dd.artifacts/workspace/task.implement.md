All 11 tasks complete. Here's a summary of changes:

## Summary

**Task 1** — `pipeline.go:34`: Added `RequireCleanTree bool` field to `Pipeline` struct with documentation comment.

**Task 2** — `pipeline.go:222`: Added `case "require_clean_tree"` in `ParsePipeline` YAML parsing, setting `p.RequireCleanTree = val == "true"`.

**Task 3** — `build.go:820-847`: Added `isWorkingTreeClean(projectRoot string) (bool, error)` helper. Shells out to `git status --porcelain`, parses output line-by-line, and ignores lines whose path starts with `.ko/`.

**Task 4** — `build.go:26-63`: Updated `BuildEligibility` signature to `BuildEligibility(ticketsDir string, t *Ticket, depsResolved bool, requireCleanTree bool)`. Adds working tree check in the `"open"` case when `requireCleanTree` is true.

**Task 5** — `build.go:70`: Updated `RunBuild` to call `BuildEligibility(ticketsDir, t, depsResolved, p.RequireCleanTree)`.

**Task 6** — `cmd_build.go`: Reordered so pipeline config is loaded before the eligibility check, then passes `p.RequireCleanTree` to `BuildEligibility`.

**Task 7** — `build_test.go`: Added `TestIsWorkingTreeClean`, `TestRequireCleanTreeRejectsDirty`, and `TestRequireCleanTreeIgnoresKoDir`, plus an `initGitRepo` helper. Required adding `os/exec` import.

**Task 8** — `pipeline_test.go`: Added `TestParseRequireCleanTree` verifying both `require_clean_tree: true` and the default-false case. Required adding `path/filepath` import.

**Task 9** — `specs/pipeline.feature`: Added two scenarios in the `# Eligibility` section — one for dirty tree blocking the build, one for `.ko/`-only changes being ignored.

**Task 10** — `testdata/pipeline/build_require_clean_tree.txtar`: Integration test that initializes a real git repo, tests that uncommitted changes outside `.ko/` block the build, and that changes only inside `.ko/` are ignored.

**Task 11** — `examples/*/pipeline.yml`: Added commented `require_clean_tree: true` line in all three example pipelines.

**No deviations from the plan.**
