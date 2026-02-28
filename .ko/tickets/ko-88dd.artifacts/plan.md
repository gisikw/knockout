## Goal
Add pipeline-level configuration option `require_clean_tree: true` to prevent build starts when uncommitted changes exist outside `.ko/`.

## Context
The build system currently allows agents to start builds regardless of working tree state. For some workflows, starting a build with uncommitted changes creates confusion about change provenance and makes it harder to attribute code to specific tickets.

Key structures:
- pipeline.go:23-41 — Pipeline struct holds top-level config (Model, MaxRetries, etc.)
- pipeline.go:131-401 — ParsePipeline reads YAML and populates Pipeline struct
- build.go:26-55 — BuildEligibility checks if a ticket can be built (status, deps)
- build.go:58-166 — RunBuild orchestrates the full build pipeline

The BuildEligibility function (build.go:26-55) is the gatekeeper that determines whether a ticket can be built. It currently checks ticket status and dependency resolution, but doesn't examine working tree state.

Working tree checks need to:
1. Run `git status --porcelain` to detect uncommitted changes
2. Ignore changes in `.ko/` directory (ticket metadata, build artifacts, etc.)
3. Return an error message if dirty changes are found and `require_clean_tree: true`

**Exception to zero-external-runtime-dependencies rule:** Shelling out to git is acceptable here because `require_clean_tree: true` is only ever invoked by the pipeline workflow, which already assumes git is available (on_succeed/on_fail hooks use git directly). This is consistent with the existing precedent where hooks shell out to user-specified commands. The ticket author confirmed this exception on 2026-02-28.

Testing pattern:
- Integration tests live in `testdata/pipeline/*.txtar` and are run by `TestPipeline` in `ko_test.go`
- Gherkin behavioral specs live in `specs/pipeline.feature`
- Both are required per INVARIANTS.md "Every behavior has a spec" and "Every spec has a test"

## Approach
Add a `RequireCleanTree` boolean field to the Pipeline struct. Parse it from YAML. Before starting a build, if this flag is true, shell out to `git status --porcelain`, filter out `.ko/` lines, and fail the build eligibility check with a clear error message if uncommitted changes exist. Add Gherkin scenarios to `specs/pipeline.feature` and corresponding `.txtar` integration tests.

## Tasks
1. [pipeline.go:23-41] — Add `RequireCleanTree bool` field to Pipeline struct
   - Place after `StepTimeout` field
   - Add comment: `// RequireCleanTree requires working tree to be clean (no uncommitted changes outside .ko/) before build starts`
   Verify: Compile succeeds.

2. [pipeline.go:131-401] — Parse `require_clean_tree` from YAML in ParsePipeline
   - Add case in top-level scalars section (around line 220)
   - Parse as boolean: `p.RequireCleanTree = val == "true"`
   Verify: Unit test that parses a pipeline with `require_clean_tree: true` and checks the field is set.

3. [build.go] — Add `isWorkingTreeClean(projectRoot string) (bool, error)` function
   - Place near other helper functions (after the `hasFlakeNix` function, currently around line 826)
   - Run `git status --porcelain` from projectRoot using `exec.Command`
   - Parse output line-by-line
   - Ignore lines where the path component (columns 4+) starts with `.ko/`
   - Return (true, nil) if no relevant changes, (false, nil) if dirty, (false, err) on git error
   Verify: Unit test with temp git repo in various states.

4. [build.go:26-55] — Add working tree check to BuildEligibility and update its signature
   - Change signature from `BuildEligibility(t *Ticket, depsResolved bool)` to `BuildEligibility(ticketsDir string, t *Ticket, depsResolved bool, requireCleanTree bool)`
   - In the "open" case, after the depsResolved check, add: if requireCleanTree is true, call isWorkingTreeClean(ProjectRoot(ticketsDir)); return an error if dirty or if git check fails
   - Error message: `"ticket '%s' cannot be built: working tree has uncommitted changes (required by require_clean_tree)"`
   Verify: Unit test that BuildEligibility rejects dirty tree when flag is set.

5. [build.go:58-65] — Update RunBuild to pass Pipeline fields to BuildEligibility
   - Change `BuildEligibility(t, depsResolved)` to `BuildEligibility(ticketsDir, t, depsResolved, p.RequireCleanTree)`
   Verify: Compile succeeds, existing tests still pass.

6. [cmd_build.go] — Update BuildEligibility call site in the build command
   - The eligibility check in cmd_build.go also calls BuildEligibility; update it to pass the new arguments (ticketsDir, t, depsResolved, p.RequireCleanTree), loading p before the check
   Verify: Compile succeeds.

7. [build_test.go] — Add unit tests for require_clean_tree enforcement
   - TestRequireCleanTreeRejectsDirty: set up temp git repo with uncommitted changes outside .ko/, call BuildEligibility with requireCleanTree=true, verify error about dirty tree
   - TestRequireCleanTreeIgnoresKoDir: set up temp git repo with uncommitted changes only inside .ko/, verify BuildEligibility allows the build
   - TestIsWorkingTreeClean: unit test for the isWorkingTreeClean helper directly (clean repo returns true, dirty repo returns false, .ko/-only changes return clean)
   Verify: New tests pass.

8. [pipeline_test.go] — Add test for parsing require_clean_tree
   - Add a TestParseRequireCleanTree test (or add a case to existing parse tests)
   - Verify `require_clean_tree: true` sets RequireCleanTree to true
   - Verify omitting the field defaults to false
   Verify: Test passes.

9. [specs/pipeline.feature] — Add Gherkin scenarios for require_clean_tree
   - Add in the "# Eligibility" section (near line 188)
   - Add two scenarios:
     - "require_clean_tree blocks build when working tree has uncommitted changes": given require_clean_tree: true in pipeline and uncommitted changes outside .ko/, when ko build runs, then the command fails with "uncommitted changes"
     - "require_clean_tree ignores changes in .ko/ directory": given require_clean_tree: true and uncommitted changes only in .ko/, when ko build runs, then build proceeds normally
   Verify: Spec reads as correct behavioral documentation.

10. [testdata/pipeline/build_require_clean_tree.txtar] — Add txtar integration test
    - Test 1 (dirty tree blocked): init git repo, make initial commit, add untracked file outside .ko/, run `ko agent build ko-a001`, verify failure with "uncommitted changes"
    - Test 2 (ko/ changes ignored): init git repo, make initial commit, modify a file in .ko/ only, run `ko agent build ko-a001`, verify build proceeds (exits with expected fake-llm output, not the clean-tree error)
    - The txtar test should use `exec git init`, `exec git config`, `exec git add .`, `exec git commit -m "init"` to set up the git state
    - Use a `require_clean_tree: true` in the pipeline.yml for the test
    Verify: `go test ./...` passes, specifically `TestPipeline` runs the new test without errors.

11. [examples/*/pipeline.yml] — Add commented example of require_clean_tree
    - In minimal, default, and structured examples
    - Add: `# require_clean_tree: true  # Prevent builds when uncommitted changes exist (outside .ko/)`
    - Place in top-level config section with other pipeline options
    Verify: Examples remain valid YAML.

## Open Questions
None. The semantic ambiguity (prevent when dirty vs. when clean) was resolved in prior build: `require_clean_tree: true` blocks builds when there ARE uncommitted changes (standard interpretation). The git dependency exception was confirmed by the ticket author on 2026-02-28. The only remaining gap from the prior failed review was the missing Gherkin spec and txtar integration test, which are now covered in tasks 9 and 10.
