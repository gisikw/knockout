Perfect! All tasks are complete. Let me provide a summary of the changes:

## Summary

I've successfully implemented all requirements from the plan. Here's what was changed:

### Task 1: Extended ticketJSON struct (cmd_query.go)
- Added three new fields to `ticketJSON`:
  - `Description string` (maps to ticket body)
  - `HasUnresolvedDep bool` (computed from dependency status)
  - `PlanQuestions []PlanQuestion` (maps to ticket plan-questions)

### Task 2: Updated cmdQuery (cmd_query.go)
- Populated the three new fields when creating ticketJSON instances
- `HasUnresolvedDep` computed using `!AllDepsResolved(ticketsDir, t.Deps)`

### Tasks 3-7: Updated all list commands (cmd_list.go)
Changed **cmdLs**, **cmdReady**, **cmdBlocked**, **cmdClosed**, and **cmdResolved** to:
- Output JSON arrays instead of JSONL (using `json.NewEncoder` with `SetIndent("", "  ")`)
- Added `--limit` flag to cmdLs, cmdReady, and cmdBlocked (cmdClosed and cmdResolved already had it)
- Populated all three new ticketJSON fields in each command
- Updated flag descriptions from "output as JSONL" to "output as JSON array"

### Task 8: Populated new fields everywhere
- All commands now properly populate `Description`, `HasUnresolvedDep`, and `PlanQuestions` when creating ticketJSON instances
- For cmdReady's cross-project path, used `AllDepsResolvedWith` for the hasUnresolvedDep computation

### Task 9: Updated spec file (specs/ticket_listing.feature)
- Changed "valid JSONL" to "valid JSON array" in 4 scenarios (ls, ready, blocked, closed)

### Task 10: Updated test file (testdata/ticket_listing/json_output.txtar)
- Updated all test assertions to expect JSON array format with proper indentation
- Added tests for `--limit` flag on ls, ready, and blocked commands
- Added test for description field when ticket has body content
- Verified hasUnresolvedDep field is correctly computed

### Verification
All tests pass (`go test ./...` ✓)

The implementation follows the plan exactly, with no deviations. The JSON output is now suitable for HTTP API responses, and the --limit flag allows pagination.
