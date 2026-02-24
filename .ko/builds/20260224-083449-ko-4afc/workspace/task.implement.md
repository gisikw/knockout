Excellent! All tests pass. Let me create a summary of the changes made.

## Summary

I have successfully implemented the `ko blocked` command enhancements as specified in the plan. Here's what was changed, organized by task:

### Task 1 & 3: Modified `cmdBlocked` in cmd_list.go:229
- **Added optional ticket ID argument**: When an ID is provided (e.g., `ko blocked <id>`), the command now resolves the ticket, loads it, and displays just that ticket's block reason.
- **Enhanced listing behavior**: When no ID is provided, `ko blocked` now lists:
  - Tickets with `status: blocked` (showing their block reason from notes)
  - Tickets with unresolved dependencies (showing the deps, as before)
- **Output format**:
  - For `status=blocked` tickets: Shows reason inline if present (e.g., `ko-d897 [blocked] (p2) Title — Reason`)
  - For dep-blocked tickets: Shows deps as before (e.g., `ko-1234 [open] (p2) Title <- [dep-id]`)
  - For specific ticket lookup: Shows `ko-1234: Reason` or `ko-1234: no block reason found`

### Task 2: Added `ExtractBlockReason` function to ticket.go:463
- **Pure function** that parses ticket body to find the most recent note containing "ko: FAIL" or "ko: BLOCKED"
- Extracts the reason text after the " — " delimiter
- Returns empty string if no block reason found
- Handles UTF-8 em dash character correctly

### Task 2 verification: Created ticket_test.go
- Comprehensive unit tests covering:
  - FAIL notes with and without reasons
  - BLOCKED notes with reasons
  - Multiple notes (returns most recent)
  - Tickets without notes sections
  - Empty bodies
- All tests pass ✓

### Task 4: Updated help text in main.go:121
- Changed from: `blocked            Show tickets with unresolved deps`
- Changed to: `blocked [id]       Show blocked tickets or reason for specific ticket`

### Task 5: Added scenarios to specs/ticket_listing.feature
- **Blocked shows tickets with status=blocked**: Verifies that tickets with `status: blocked` appear in the list
- **Blocked shows block reason for status=blocked tickets**: Verifies that block reasons are displayed inline
- **Blocked with ID shows specific ticket's block reason**: Tests the new `ko blocked <id>` command
- **Blocked with ID shows message when no reason found**: Tests the "no block reason found" case

### Verification
- ✓ All existing tests pass
- ✓ New unit test for `ExtractBlockReason` passes
- ✓ Manual testing confirms both commands work:
  - `ko blocked` lists all blocked tickets with reasons
  - `ko blocked <id>` shows specific ticket's block reason
  - Partial ID resolution works (e.g., `ko blocked f65e`)

### No deviations from plan
All tasks were completed as specified. The implementation follows existing patterns in the codebase and complies with INVARIANTS.md.
