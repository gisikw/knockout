# Implementation Summary: Consolidate Project Commands

## What Was Done

Successfully consolidated four separate project management commands (`init`, `register`, `default`, `projects`) into two unified subcommands under `ko project`:

- `ko project set #<tag> [--prefix=p] [--default]` - Upsert operation that:
  - Initializes `.ko/tickets/` directory if needed
  - Writes prefix to `.ko/config.yaml` (if provided)
  - Registers project in global registry (`~/.config/knockout/projects.yml`)
  - Optionally sets project as default
  - Detects prefix from existing tickets if not explicitly provided

- `ko project ls` - Lists all registered projects with default marker (`*`)

## Files Changed

**Created:**
- `cmd_project.go` - New command dispatcher with `cmdProject`, `cmdProjectSet`, `cmdProjectLs`
- `cmd_project_test.go` - Comprehensive test suite (12 tests covering all behaviors)

**Deleted:**
- `cmd_init.go` - Replaced by `ko project set` with `--prefix` flag
- `cmd_init_test.go` - Functionality covered by new tests
- `cmd_registry.go` - Old command implementations removed

**Modified:**
- `main.go` - Updated dispatcher to route `project` command, removed old command cases, updated help text
- `registry.go` - Added `findProjectRoot()` helper (preserved from deleted cmd_registry.go)
- `README.md` - Updated usage examples and project registry documentation
- `specs/project_registry.feature` - Updated Gherkin specs to reflect new command structure
- `testdata/project_registry/*.txtar` - Updated 12 txtar test files to use new commands

## Notable Implementation Decisions

1. **Upsert Semantics**: `project set` is idempotent - running it multiple times on the same project updates the registration. This provides a simpler mental model than separate init/register/default commands.

2. **Prefix Detection**: When `--prefix` is omitted, the system attempts to detect the prefix from existing ticket files in `.ko/tickets/`. This preserves existing behavior while simplifying the common case.

3. **Success Messages**: Output messages vary based on which operations were performed (init, register, set default) to provide clear feedback to the user.

4. **Default Visibility**: Instead of requiring a separate `ko default` command to query the default project, `ko project ls` shows it inline with an asterisk marker, following common CLI conventions (e.g., git branch).

5. **Spec Compliance**: Updated the Gherkin spec file to maintain the invariant that "every behavior has a spec" - the spec is the source of truth, not the implementation.

## Test Coverage

All planned tests implemented and passing:
- Subcommand validation (no args, unknown subcommand)
- Minimal registration (tag only)
- Registration with prefix
- Registration with default flag
- Upsert behavior (updating existing projects)
- Directory creation (.ko/tickets)
- Error cases (no tag, invalid prefix length)
- Empty registry listing
- Multiple project listing with default marker

Integration: All 12 existing txtar tests updated and passing.

## Compliance Notes

- **INVARIANTS.md**: Complies with all relevant invariants:
  - Specs updated before merging (project_registry.feature)
  - Tests mirror source files (cmd_project_test.go)
  - Decision logic is pure (argument parsing, validation)
  - File size: cmd_project.go is 170 lines (well under 500 line limit)

- **No Scope Creep**: Implementation stayed strictly within ticket requirements - no unrelated refactoring or feature additions.

## Future Considerations

None - the implementation is complete and self-contained. The old commands are fully removed, and all functionality is preserved through the new unified interface.
