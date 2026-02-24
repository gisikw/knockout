Perfect! All tests pass. Let me create a summary of what was implemented:

## Summary

Successfully implemented `--json` flag across all output-producing ko commands. All 9 tasks from the plan were completed without deviations.

### Changes by Task:

**Task 1-4: List Commands (cmd_list.go)**
- Added `--json` flag to `cmdLs`, `cmdReady`, `cmdBlocked`, and `cmdClosed`
- When `--json` is set, each command outputs JSONL (newline-delimited JSON) using the `ticketJSON` struct
- Reused the existing `ticketJSON` struct from `cmd_query.go` for consistency
- Added required imports: `encoding/json` and `time`

**Task 5: Show Command (cmd_show.go)**
- Added `--json` flag to `cmdShow`
- Created new `showJSON` struct that extends `ticketJSON` with derived fields:
  - `blockers` - array of open dependencies
  - `blocking` - array of tickets depending on this one
  - `children` - array of child ticket IDs
  - `body` - ticket body content
- JSON output is pretty-printed with indentation for readability
- Added required imports: `encoding/json`, `flag`, and `time`

**Task 6-7: Specifications**
- Added 4 scenarios to `specs/ticket_listing.feature` covering `--json` for ls, ready, blocked, and closed commands
- Added 1 comprehensive scenario to `specs/ticket_show.feature` covering `--json` with all derived fields

**Task 8-9: Tests**
- Created `testdata/ticket_listing/json_output.txtar` with tests for all list commands
- Created `testdata/ticket_show/json_output.txtar` with tests for show command including derived fields

### Verification:
- All tests pass (`go test ./...`)
- Implementation follows existing patterns from `ko query`
- Complies with INVARIANTS.md (specs before code, tests for every spec)
- Minimal changes - only added what was required by the plan
