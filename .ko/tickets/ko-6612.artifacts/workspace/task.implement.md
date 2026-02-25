Excellent! The auto-unblock functionality is working correctly. The ticket was blocked when questions were added, and then automatically unblocked when all questions were answered.

## Summary

I've successfully implemented the `ko update` command according to the plan. Here's what was completed:

### Task 1: Create cmd_update.go ✓
- Created `cmd_update.go` with comprehensive flag parsing for all mutable ticket fields
- Implemented support for: `--title`, `-d` (description), `-t` (type), `-p` (priority), `-a` (assignee), `--parent`, `--external-ref`, `--design`, `--acceptance`, `--tags`, `--questions`, `--answers`, `--status`
- Tags replace (not append) as specified
- Questions handling sets status to blocked
- Answers handling includes auto-unblock logic: when all questions are answered, status transitions from `blocked` to `open`
- Proper error handling for invalid inputs
- Emits mutation events for observability

### Task 2: Add update command to main.go ✓
- Added `case "update": return cmdUpdate(rest)` to the switch statement in main.go
- Placed alphabetically after `undep` as specified

### Task 3: Add help text in cmdHelp() ✓
- Added comprehensive help text showing all available flags
- Placed after triage block and before dep block as specified
- Includes note about tags replacing and answers auto-unblocking

### Task 4: Create cmd_update_test.go ✓
- Created comprehensive test suite with 6 test functions covering:
  - Basic field updates (title, type, priority, assignee)
  - Tags replacement (not append)
  - Status updates and validation
  - Auto-unblock behavior (partial and full answer scenarios)
  - Error cases (missing ID, nonexistent ticket, no fields, invalid JSON)
  - Description, design, and acceptance criteria updates
- All tests pass

### Verification:
- `go build .` compiles without errors
- `go test ./...` passes all tests (existing + new)
- `ko help` shows the update command with correct usage
- Functional tests confirm:
  - Basic updates work (title, type, priority, tags)
  - Tags replace existing tags (not append)
  - Questions block the ticket
  - Answering all questions auto-unblocks the ticket (status: blocked → open)

The implementation follows all existing patterns in the codebase and complies with the INVARIANTS.md requirements.
