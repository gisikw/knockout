# Implementation Summary: Resolved Status

## What Was Done

Successfully added a new `resolved` status to the ticket system for agent-completed work that requires human review before closing.

### Changes Made

1. **ticket.go:15** — Added "resolved" to the Statuses slice
2. **ticket.go:66** — Updated IsReady to exclude "resolved" status (grouped with "closed")
3. **ticket.go:422** — Updated statusOrder to position resolved (rank 2) between open/in_progress and closed
4. **cmd_list.go:401-470** — Added cmdResolved function with JSON and text output support
5. **main.go:46** — Added "resolved" case to command dispatcher
6. **main.go:135** — Added "resolved" to help text
7. **ticket_test.go:5-63** — Added comprehensive IsReady unit tests covering all statuses
8. **INVARIANTS.md:54-58** — Updated invariants documentation to reflect the new status

## Implementation Decisions

- **Placement in status order**: Resolved (rank 2) sits between active statuses (open/in_progress at 0-1) and closed (rank 3), with blocked/other statuses at rank 4. This ensures resolved tickets appear after active work but before fully closed tickets in sorted views.

- **Ready queue exclusion**: Resolved tickets are treated like closed tickets — they're excluded from the ready queue because they represent completed work awaiting review, not actionable work.

- **Command parity**: The `ko resolved` command mirrors the structure of `ko closed`, supporting both text and JSON output modes with optional limit flags.

- **Test coverage**: Added 7 test cases covering all status/dependency combinations for IsReady, confirming resolved tickets are never ready regardless of dependency state.

## Verification

- All existing tests pass: `go test ./... -count=1`
- New tests verify IsReady behavior for resolved status
- Manual verification:
  - `ko status <id> resolved` successfully sets status
  - `ko resolved` lists resolved tickets
  - `ko ready` excludes resolved tickets
  - Help text displays correctly

## Notes for Future Work

As noted in the plan, this ticket only adds the status to the data model. Workflows that want to transition tickets to `resolved` instead of `closed` will need prompt updates or custom disposition handling in their decision nodes. The current build.go outcome logic continues to use `closed` for successful completions.
