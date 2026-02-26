# Summary: #tag Shorthand for --project Flag

## What Was Done

Successfully implemented `#tag` as shorthand syntax for `--project=tag` across all commands that support project filtering (e.g., `ko ls #punchlist` now works as an alternative to `ko ls --project=punchlist`).

### Changes Made

1. **Extended `resolveProjectTicketsDir()` function (cmd_list.go:54-100)**
   - Added two-pass argument parsing:
     - First pass: scan for positional `#tag` arguments and extract the project tag
     - Second pass: check for explicit `--project` flag (which overrides `#tag`)
   - Used existing `CleanTag()` function to strip the leading `#` character
   - Implemented precedence logic: explicit `--project` flag takes precedence over implicit `#tag` shorthand
   - Updated function documentation to reflect new behavior
   - Consumed `#tag` arguments from the remaining args list (don't pass through)

2. **Added comprehensive test coverage (cmd_list_test.go:131-275)**
   - `TestResolveProjectTicketsDir_HashTagShorthand`: Verifies `#tag` at beginning of args
   - `TestResolveProjectTicketsDir_HashTagAnyPosition`: Verifies `#tag` works regardless of position
   - `TestResolveProjectTicketsDir_HashTagUnknownProject`: Verifies proper error handling for unknown projects
   - `TestResolveProjectTicketsDir_ProjectFlagOverridesHashTag`: Verifies precedence rules

3. **Added behavioral specification (specs/project_registry.feature:87-92)**
   - New scenario: "Commands accept #tag shorthand for --project flag"
   - Documents expected behavior using gherkin syntax
   - Example: `ko ls #exo` should list tickets from the exo project

### Verification
- ✅ All tests pass (`go test ./...`)
- ✅ All four planned tasks completed
- ✅ No invariant violations detected
- ✅ Backward compatible: all existing `--project` syntax still works
- ✅ No unrelated changes made

## Implementation Decisions

### Precedence Rules
When both `#tag` and `--project` are provided, `--project` takes precedence. This follows the principle that explicit flags should override implicit shorthands. Example:
```bash
ko ls #proj1 --project=proj2  # Uses proj2, not proj1
```

### Only First #tag Consumed
If multiple `#tag` arguments are present, only the first one is consumed as the project tag. This prevents ambiguity and follows the principle of deterministic behavior. Example:
```bash
ko ls #proj1 #proj2  # Uses proj1, #proj2 is passed through as a regular arg
```

### Two-Pass Parsing
The implementation uses two-pass parsing to correctly handle the precedence:
1. First pass scans all args for `#tag` patterns
2. Second pass checks for `--project` flags (which override `#tag`)

This ensures that `--project` always wins, regardless of the order of arguments.

## Impact

This change automatically enables `#tag` shorthand across all commands that call `resolveProjectTicketsDir()`:
- `ko ls #tag`
- `ko ready #tag`
- `ko show #tag`
- `ko update #tag`
- `ko dep #tag` / `ko undep #tag`
- `ko note #tag`
- `ko bump #tag`
- `ko agent build/loop/start/stop/status #tag`
- `ko agent init #tag`

No changes were needed to individual commands themselves, demonstrating good architectural separation.

## Notes for Future Readers

1. The `CleanTag()` function (registry.go:148-151) is the canonical way to strip leading `#` from tags. Any new code that needs to handle `#tag` syntax should use this function.

2. The existing `ko project set #tag` command already used this pattern, so we followed the established convention.

3. The behavioral spec (specs/project_registry.feature) documents the behavior but is not currently executable. Per INVARIANTS.md, every spec should have a corresponding test. The unit tests in cmd_list_test.go serve as the verification layer.

4. If additional commands are added that need to resolve project tags, they should use `resolveProjectTicketsDir()` to automatically inherit the `#tag` shorthand support.

## Compliance Check

- ✅ **Spec before code**: Behavioral spec added (specs/project_registry.feature)
- ✅ **Every spec has a test**: Unit tests added (cmd_list_test.go)
- ✅ **No invariant violations**: All INVARIANTS.md requirements satisfied
- ✅ **Tests pass**: `go test ./...` exits successfully
- ✅ **Stay within scope**: No unrelated changes made
- ✅ **No over-engineering**: Simple implementation using existing infrastructure
