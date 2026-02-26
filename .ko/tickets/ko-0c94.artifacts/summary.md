# After-Action Summary: Remove ko triage and ko blocked commands

## What Was Done

Successfully removed the `ko triage` and `ko blocked` commands from the knockout codebase as they have been superseded by `ko update`. This was a clean deletion with no architectural changes.

### Files Deleted
- `cmd_triage.go` - 290 lines of triage command implementation
- `cmd_triage_test.go` - 828 lines of comprehensive test coverage

### Files Modified
- **cmd_list.go**: Removed `cmdBlocked` function (123 lines)
- **main.go**: Removed switch cases for "triage" and "blocked" commands, removed help text documentation
- **ticket.go**: Removed `ExtractBlockReason` helper function (44 lines) - only used by removed commands
- **ticket_test.go**: Removed `TestExtractBlockReason` test function (55 lines)

### Test Data Updated
Updated 4 txtar test files to replace removed commands with equivalent functionality:
- `status_ready_exclusion.txtar`: Changed `ko triage --block` to `ko block`
- `status_shortcuts.txtar`: Changed `ko triage --block` to `ko block`
- `blocked_view.txtar`: Replaced `ko blocked` with `ko ls` and `ko ready`
- `json_output.txtar`: Replaced `ko blocked --json` with `ko ls --json`

## Implementation Decisions

1. **Clean removal**: `ExtractBlockReason` was removed because it was only used by the two commands being deleted. No other code depended on it.

2. **Test updates**: Rather than deleting the behavioral tests, they were updated to use equivalent commands (`ko block`, `ko ls`, `ko ready`), preserving test coverage of the underlying functionality.

3. **No changes to `ko update`**: The superseding command already has all needed functionality via `--questions` and `--answers` flags - no changes required.

## Verification

- ✅ All tests pass (`go test ./...`)
- ✅ Project builds successfully (`go build`)
- ✅ No references to `cmdTriage` or `cmdBlocked` remain in source code
- ✅ No references to `ExtractBlockReason` remain in source code
- ✅ All INVARIANTS.md contracts upheld

## Notes for Future Readers

The `ko triage` command provided functionality for blocking tickets and managing questions. This is now handled by:
- `ko block <id> [reason]` - for blocking tickets
- `ko update <id> --questions '<json>'` - for adding questions
- `ko update <id> --answers '<json>'` - for answering questions

The `ko blocked` command showed blocked tickets. This is now handled by:
- `ko ls --status blocked` - list blocked tickets
- `ko ls` - shows tickets with unresolved deps

No migration is needed as these are CLI commands, not data formats.
