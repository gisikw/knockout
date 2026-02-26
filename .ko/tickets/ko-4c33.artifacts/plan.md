## Goal
Support `#tag` as shorthand syntax for `--project=tag` in commands that accept the `--project` flag.

## Context
The codebase already has infrastructure to support `#tag` syntax in specific places:
- `cmd_project.go:54` uses `CleanTag()` to strip leading `#` from tags
- `registry.go:148-151` defines `CleanTag()` to strip one leading `#`
- The README shows `ko agent start '#myapp'` as an example (line 196-198)
- All cross-project commands use `resolveProjectTicketsDir()` which currently only handles `--project=tag` (not `#tag`)

Commands that call `resolveProjectTicketsDir()` and would benefit from `#tag` shorthand:
- `ko ls` (cmd_list.go)
- `ko ready` (cmd_list.go)
- `ko show` (cmd_show.go)
- `ko update` (cmd_update.go)
- `ko dep` / `ko undep` (cmd_dep.go)
- `ko note` (cmd_note.go)
- `ko bump` (cmd_bump.go)
- `ko agent build/loop/start/stop/status` (cmd_agent.go)
- `ko agent init` (cmd_build_init.go)

The function `resolveProjectTicketsDir()` (cmd_list.go:54-100) currently:
1. Manually parses `--project=value` and `--project value` from args
2. Extracts the project tag and looks it up in the registry
3. Returns the tickets directory and remaining args (with project flag removed)

## Approach
Extend `resolveProjectTicketsDir()` to recognize positional arguments starting with `#` as project tags. When encountering `#tag` in args, treat it as equivalent to `--project=tag` by stripping the `#` and using the tag for registry lookup. This maintains the existing API while adding convenient shorthand syntax.

## Tasks
1. [cmd_list.go:resolveProjectTicketsDir] — Extend argument parsing to detect positional args starting with `#` and treat them as project tags.
   - Before the existing flag parsing loop, scan for positional `#tag` args
   - When found, strip the `#` using `CleanTag()` and set `projectTag` variable
   - Remove the `#tag` arg from the args list (don't include in `remaining`)
   - If both `#tag` and `--project` are provided, `--project` takes precedence (explicit beats implicit)
   Verify: `go test ./... -run TestResolveProjectTicketsDir` passes.

2. [cmd_list_test.go] — Add test cases for `#tag` shorthand syntax.
   - Test case: `resolveProjectTicketsDir([]string{"#testproj", "arg1"})` resolves to testproj's tickets dir
   - Test case: `resolveProjectTicketsDir([]string{"arg1", "#testproj"})` works regardless of position
   - Test case: `resolveProjectTicketsDir([]string{"#unknown"})` returns "unknown project" error
   - Test case: `resolveProjectTicketsDir([]string{"#testproj", "--project=other"})` prefers `--project` flag
   Verify: `go test ./... -run TestResolveProjectTicketsDir` passes with new tests.

3. [specs/project_registry.feature] — Add behavioral spec scenario for `#tag` shorthand.
   - Add scenario under "# Listing" section (around line 86):
     ```gherkin
     Scenario: Commands accept #tag shorthand for --project flag
       Given a registry with project "exo" at "/tmp/test-projects/exo"
       And project "exo" has a ticket "exo-0001" with title "Test ticket"
       When I run "ko ls #exo"
       Then the command should succeed
       And the output should contain "exo-0001"
     ```
   Verify: Spec documents the new behavior.

4. [cmd_project_test.go] — Verify `ko project set` already handles `#tag` (no changes needed).
   - Examine existing tests (lines 36, 68, 100, 127, 169, 198)
   - Confirm all use `#tag` syntax and pass `CleanTag()` correctly
   Verify: `go test ./... -run TestProjectSet` passes (no code changes).

## Open Questions
None. The implementation is straightforward:
- Use existing `CleanTag()` function for consistency
- Follow existing pattern from `cmd_project.go:53-54` for detecting and stripping `#`
- Maintain backward compatibility by keeping all existing `--project` syntax working
- No changes needed to commands themselves, only to the shared `resolveProjectTicketsDir()` function
