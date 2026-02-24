## Goal
Add a `resolved` status for agent-completed work that requires human review before closing.

## Context
The ticket system uses a closed set of statuses defined in `ticket.go:15`:
```
var Statuses = []string{"captured", "routed", "open", "in_progress", "closed", "blocked"}
```

The `IsReady` function (ticket.go:62-69) determines which tickets appear in the ready queue — currently only `open` and `in_progress` tickets with resolved deps.

The `ko ready` command (cmd_list.go:123-227) queries the ready queue. The `ko closed` command (cmd_list.go:332-399) lists closed tickets.

Build outcomes in `build.go` close tickets on success (line 132: `setStatus(ticketsDir, t, "closed")`). The agent loop never creates a `resolved` status.

The pipeline workflows (`task`, `research`, `bug`) all end by closing tickets. Research and bug wontfix workflows would benefit from transitioning to `resolved` instead of `closed` for human verification.

Testing pattern: Pure functions get unit tests. See `ticket_test.go` for examples.

## Approach
1. Add `resolved` to the `Statuses` slice in ticket.go
2. Update `IsReady` to exclude `resolved` status from ready queue (like `closed`)
3. Add a `ko resolved` command that lists tickets with status=resolved
4. Update `statusOrder` in ticket.go to position resolved between in_progress/open and closed for sort ordering

The build.go outcome logic stays unchanged — workflows that want to use `resolved` can add a decision node at the end that returns a custom disposition. This is outside the scope of this ticket.

## Tasks
1. [ticket.go:15] — Add "resolved" to the Statuses slice after "blocked".
   Verify: `go test ./... -count=1` passes, ValidStatus("resolved") returns true.

2. [ticket.go:62-69 IsReady] — Update the switch statement to exclude "resolved" from the ready queue, treating it like "closed".
   Verify: Write a unit test confirming IsReady returns false for resolved status.

3. [ticket.go:414-425 statusOrder] — Add a case for "resolved" that returns 2 (between open/in_progress and closed), shifting closed to 3 and default to 4.
   Verify: Existing sort tests pass, tickets sort correctly with resolved status.

4. [cmd_list.go] — Add a `cmdResolved` function similar to `cmdClosed` that filters for status="resolved".
   Verify: Function compiles, returns appropriate exit codes.

5. [main.go:23-81] — Add case for "resolved" command that calls cmdResolved.
   Verify: `ko resolved` works from CLI.

6. [main.go:122-166 cmdHelp] — Add "resolved" command to help text after "blocked" line.
   Verify: Help text displays correctly.

7. [ticket_test.go] — Add unit test for IsReady with resolved status.
   Verify: Test passes, confirms resolved tickets don't appear in ready queue.

## Open Questions
None. This is a pure data model change. The implementation is straightforward: add the status to the allowed list, exclude it from ready queue logic, add a listing command. The workflow-level decision of when to use `resolved` instead of `closed` is a separate concern handled by prompt updates or disposition customization, not this ticket.
