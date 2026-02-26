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

Similar patterns in the codebase:
- build.go:690-727 — runHooks expands env vars and executes shell commands
- The project already shells out to git for operations (visible in on_succeed/on_fail hooks)

## Approach
Add a `RequireCleanTree` boolean field to the Pipeline struct. Parse it from YAML. Before starting a build, if this flag is true, check the git working tree status (ignoring `.ko/`). If uncommitted changes exist, fail the build eligibility check with a clear error message.

The working tree check will be added as a new function that can be called from BuildEligibility. It will shell out to `git status --porcelain`, filter out lines starting with `.ko/`, and return whether the tree is clean.

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
   - Place near other helper functions (after line 799)
   - Run `git status --porcelain` from projectRoot
   - Parse output line-by-line
   - Ignore lines where path starts with `.ko/`
   - Return (true, nil) if no relevant changes, (false, nil) if dirty, (false, err) on git error
   Verify: Unit test with temp git repo in various states.

4. [build.go:26-55] — Add working tree check to BuildEligibility
   - After line 54 (before final return "")
   - Add new check:
     ```
     if requireCleanTree {
       // Need projectRoot — calculate from ticketsDir
       projectRoot := ProjectRoot(ticketsDir)
       clean, err := isWorkingTreeClean(projectRoot)
       if err != nil {
         return fmt.Sprintf("ticket '%s' cannot be built: failed to check working tree status: %v", t.ID, err)
       }
       if !clean {
         return fmt.Sprintf("ticket '%s' cannot be built: working tree has uncommitted changes (required by require_clean_tree)", t.ID)
       }
     }
     ```
   - Note: BuildEligibility needs to accept a Pipeline parameter to access RequireCleanTree
   Verify: Unit test that BuildEligibility rejects dirty tree when flag is set.

5. [build.go:58] — Update RunBuild to pass Pipeline to BuildEligibility
   - Line 61: Change `BuildEligibility(t, depsResolved)` to `BuildEligibility(ticketsDir, t, depsResolved, p.RequireCleanTree)`
   - Update function signature at line 26-28
   Verify: Compile succeeds, existing tests still pass.

6. [build_test.go] — Add test for require_clean_tree enforcement
   - Create test: TestRequireCleanTreeRejectsDirty
   - Set up temp git repo with uncommitted changes outside .ko/
   - Create pipeline with RequireCleanTree: true
   - Call BuildEligibility and verify it returns error about dirty tree
   - Create second test: TestRequireCleanTreeIgnoresKoDir
   - Set up temp git repo with uncommitted changes only in .ko/
   - Verify BuildEligibility allows the build
   Verify: Tests pass.

7. [pipeline_test.go] — Add test for parsing require_clean_tree
   - Add test case to existing parse tests
   - Verify field is correctly parsed from YAML
   Verify: Test passes.

8. [examples/*/pipeline.yml] — Add commented example of require_clean_tree
   - In minimal, default, and structured examples
   - Add comment: `# require_clean_tree: true  # Prevent builds when uncommitted changes exist (outside .ko/)`
   - Place in top-level config section with other pipeline options
   Verify: Examples remain valid YAML.

## Open Questions
**Ticket wording inconsistency**: The ticket title says "prevent an agent from starting if the working tree is **clean**" but the config name `require_clean_tree: true` and Git conventions suggest it should prevent starting if the tree is **dirty** (i.e., has uncommitted changes). I'm proceeding with the standard interpretation: `require_clean_tree: true` means "require the tree to be clean" → prevent builds when there are uncommitted changes. If the ticket actually wants the opposite behavior (prevent when clean, allow when dirty), that would be unusual and should be clarified.
