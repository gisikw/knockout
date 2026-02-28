# After-Action Summary

## What Was Done

Added `require_clean_tree: true` support to `pipeline.yml`. When set, a build is blocked before it starts if the git working tree has uncommitted changes outside the `.ko/` directory.

### Changes

1. **`pipeline.go`** — Added `RequireCleanTree bool` field to `Pipeline` struct; added `case "require_clean_tree"` to the YAML parser.

2. **`build.go`** — Added `isWorkingTreeClean(projectRoot string) (bool, error)` helper that runs `git status --porcelain`, parses output line-by-line, and ignores lines whose path starts with `.ko/`. Updated `BuildEligibility` signature to accept `ticketsDir string` and `requireCleanTree bool`; added the clean-tree gate in the `"open"` case.

3. **`cmd_build.go`** — Moved the eligibility check to after pipeline loading (so `p.RequireCleanTree` is available), updated the `BuildEligibility` call with the new arguments.

4. **`build_test.go`** — Added `TestIsWorkingTreeClean`, `TestRequireCleanTreeRejectsDirty`, and `TestRequireCleanTreeIgnoresKoDir`.

5. **`pipeline_test.go`** — Added `TestParseRequireCleanTree` covering both `true` and omitted (defaults to `false`).

6. **`specs/pipeline.feature`** — Added two Gherkin scenarios in the Eligibility section: one blocking dirty tree, one ignoring `.ko/`-only changes.

7. **`testdata/pipeline/build_require_clean_tree.txtar`** — Integration test covering both scenarios end-to-end; passes as part of `TestPipeline`.

8. **`examples/*/pipeline.yml`** — Added commented-out `require_clean_tree` option in all three example pipelines.

## Notable Decisions

- **Git dependency exception confirmed.** Shelling out to `git` is acceptable here because `require_clean_tree` runs in the same context as `on_succeed`/`on_fail` hooks, which already assume git is available. This was confirmed by the ticket author on 2026-02-28 and documented in the plan.

- **`cmd_build.go` reordering.** The eligibility check was moved to after pipeline loading. This is a minor structural improvement with no behavioral change for the existing non-`require_clean_tree` path, since eligibility was previously checked before pipeline was loaded (meaning `p.RequireCleanTree` was unavailable). The new order is: load ticket → load pipeline → check eligibility → run build.

- **Rename target handling.** `git status --porcelain` emits `old -> new` for renames. The parser strips the rename arrow and checks only the destination path against `.ko/`.

## For Future Readers

- The `isWorkingTreeClean` function is a pure-ish helper: it shells out to git but makes no state changes. It belongs in `build.go` near `hasFlakeNix` (the other git-adjacent helper).
- If `.ko/` is ever renamed, this check will need updating in two places: `isWorkingTreeClean` in `build.go` and the txtar test's `.ko/` path prefix.
