## Goal
Consolidate four commands (`init`, `register`, `default`, `projects`) into two new subcommands (`project set`, `project ls`).

## Context
The codebase currently has four separate top-level commands for project management:
- `cmd_init.go:cmdInit` - initializes `.ko` directory with a prefix
- `cmd_registry.go:cmdRegister` - registers current project in global registry
- `cmd_registry.go:cmdDefault` - shows or sets the default project
- `cmd_registry.go:cmdProjects` - lists all registered projects

Key structures and functions:
- `registry.go:Registry` - holds Default, Projects, Prefixes maps
- `registry.go:LoadRegistry/SaveRegistry` - I/O for `~/.config/knockout/projects.yml`
- `cmd_init.go:WriteConfigPrefix` - writes prefix to `.ko/config.yaml`
- `cmd_create.go:ReadPrefix` - reads prefix from config or legacy `.ko/prefix`
- `registry.go:detectPrefixFromDir` - auto-detects prefix from existing ticket files

Testing patterns:
- Unit tests in `*_test.go` files mirror source files
- Tests use `t.TempDir()` for isolated filesystem operations
- Registry tests exist in `registry_test.go`
- Init tests exist in `cmd_init_test.go`

Main dispatcher in `main.go:run()` uses a switch statement for command routing.

## Approach
Create a new `cmd_project.go` file with a dispatcher `cmdProject` that handles two subcommands: `set` and `ls`. The `set` subcommand will be an upsert operation combining init+register+default behavior. The `ls` subcommand will replace both `projects` (list all) and `default` (show default) commands. Update the main dispatcher to route "project" commands to the new handler. Preserve existing helper functions from `registry.go` and `cmd_init.go`. Add comprehensive tests in `cmd_project_test.go`. Remove old command implementations and tests.

## Tasks
1. [cmd_project.go] — Create new file with `cmdProject(args []string)` dispatcher that handles "set" and "ls" subcommands.
   - `set` subcommand: Parse #tag and optional --prefix, --default flags. Validate tag is provided and prefix (if given) is ≥2 chars. Find or use current directory as project root. Create `.ko/tickets/` if needed. Write prefix to config.yaml (init). Register in global registry (register). Optionally set as default if --default flag present. Print appropriate success message.
   - `ls` subcommand: Load registry, list all projects sorted by tag, show asterisk marker for default project. Handle empty registry gracefully.
   - Error handling: proper stderr messages and non-zero exit codes for all failure cases.
   Verify: `go build` compiles.

2. [cmd_project_test.go] — Create comprehensive test suite covering all behaviors.
   - Test `set` with just tag (minimal case)
   - Test `set` with tag and --prefix
   - Test `set` with --default flag
   - Test `set` updates existing project (upsert semantics)
   - Test `set` creates .ko directory if missing
   - Test `set` error cases: no args, invalid prefix length, can't write config
   - Test `ls` with empty registry
   - Test `ls` with multiple projects
   - Test `ls` shows default marker correctly
   - Test `project` with no subcommand shows usage
   - Test `project` with unknown subcommand shows error
   Verify: `go test -run TestCmdProject` passes.

3. [main.go:run] — Update command dispatcher switch statement.
   - Add `case "project": return cmdProject(rest)` before the default case
   - Remove cases for "init", "register", "default", "projects"
   Verify: `go build` compiles.

4. [main.go:cmdHelp] — Update help text to reflect new commands.
   - Remove lines for `init`, `register`, `default`, `projects`
   - Add new section for project management:
     ```
     project set #<tag> [--prefix=p] [--default]
                                Initialize .ko dir, register project, optionally set default
     project ls                 List registered projects (default marked with *)
     ```
   Verify: `ko help` output shows new commands, old commands removed.

5. [cmd_init.go, cmd_init_test.go] — Delete both files entirely.
   Verify: `go build` compiles, old tests no longer run.

6. [cmd_registry.go] — Remove `cmdRegister`, `cmdDefault`, `cmdProjects`, `findProjectRoot` functions.
   - Keep: `Registry` type, `LoadRegistry`, `SaveRegistry`, `ParseRegistry`, `FormatRegistry`, `CrossProjectLookup`, `extractPrefix`, `RoutingDecision`, `ParseTags`, `RouteTicket`, `CleanTag`, helper functions.
   - Remove: The four cmd* functions (lines 8-154).
   Verify: `go build` compiles, `go test ./...` passes.

7. [registry_test.go] — Review and keep all existing tests since they test pure functions that remain.
   Verify: `go test -run TestParseRegistry` and other registry tests pass.

8. Integration smoke test — Manually verify end-to-end behavior.
   - `ko project set #test --prefix=tst` in a temp directory creates .ko, registers project
   - `ko project ls` shows the registered project
   - `ko project set #test --default` marks it as default
   - `ko project ls` shows asterisk marker
   - Old commands `ko init`, `ko register`, etc. return unknown command error
   Verify: All manual tests pass, help text correct, no regressions.

## Open Questions
None — the ticket clearly specifies the API and the existing code provides all necessary building blocks. The implementation is straightforward: consolidate four commands into a two-subcommand structure with upsert semantics for `set` and unified listing for `ls`.
