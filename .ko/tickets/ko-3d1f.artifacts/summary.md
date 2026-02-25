# Implementation Summary: ko questions command

## What Was Done

Added a new `ko questions <id>` command that reads the `plan-questions` field from ticket frontmatter and outputs it as JSON to stdout.

### Files Created
- `cmd_questions.go` - Main command implementation (43 lines)
- `cmd_questions_test.go` - Comprehensive unit tests (349 lines)

### Files Modified
- `main.go` - Added command routing and help text
- `specs/ticket_status.feature` - Added behavioral specifications for the new command

## Implementation Details

The command follows the established pattern from `cmd_query.go` and `cmd_answer.go`:

1. **Command structure**: Uses `resolveProjectTicketsDir()` for cross-project routing and `ResolveID()` for partial ID matching
2. **JSON output**: Uses `json.NewEncoder(os.Stdout)` with indentation for readability
3. **Error handling**: Returns proper exit codes and writes errors to stderr
4. **Edge cases**: Handles nil/empty plan-questions by outputting `[]`

### Test Coverage

Created three test functions covering:
- Tickets with plan-questions (outputs full JSON array)
- Tickets without plan-questions (outputs empty array `[]`)
- Tickets with nil plan-questions (outputs `null`)
- Error cases (missing ID, nonexistent ticket)
- Partial ID resolution

All tests pass: `go test -run TestCmdQuestions` succeeds.

## Notable Decisions

**Spec compliance**: Added behavioral scenarios to `specs/ticket_status.feature` to comply with INVARIANTS.md requirement that "Every behavior has a spec." The dependency tickets (ko-9c4a, ko-16e8) only included unit tests, but INVARIANTS.md states "No grandfathering" for compliance issues, so specs were added during review.

**JSON formatting**: Used `SetIndent("", "  ")` for human-readable output, consistent with the ticket's example output showing indented JSON.

**Nil vs empty array**: The implementation correctly handles both `nil` and empty `[]` slices in `t.PlanQuestions`, outputting JSON appropriately for each case.

## Verification

- ✅ Command compiles without errors
- ✅ All unit tests pass
- ✅ Help text displays correctly
- ✅ Integration with existing codebase (uses shared ticket loading, ID resolution)
- ✅ Behavioral specifications added to spec file

## Dependencies

This command completes the read/write/update trio for plan-questions:
- ko-9c4a: `ko block --questions` (write)
- ko-16e8: `ko answer` (update/consume)
- ko-3d1f: `ko questions` (read) ← this ticket

The punchlist server can now use `ko questions` to fetch questions for rendering UI on blocked tickets.
