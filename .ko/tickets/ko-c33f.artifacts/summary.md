# Summary: ko-c33f — Exclude snoozed tickets from `ko ready`

## What was done

Added snooze-awareness to `ko ready` and the loop's `ReadyQueue`. Tickets with a future `snooze` date are now excluded from the ready queue.

**Changes:**
- `ticket.go`: Added pure `IsSnoozed(snooze string, now time.Time) bool`. Parses the date in UTC via `time.ParseInLocation`; returns `false` for empty/invalid strings (safe no-op per plan). Returns `now.Before(parsedDate)` where `parsedDate` is midnight UTC on the snooze day — correctly implementing "valid as of midnight on that date" semantics.
- `ticket_test.go`: Added `TestIsSnoozed` with 5 deterministic cases: empty, past, future, today (snoozed date == today → not snoozed), and invalid string.
- `cmd_list.go`: Both `IsReady(...)` call sites in `cmdReady` now also gate on `!IsSnoozed(t.Snooze, time.Now())`.
- `loop.go`: `ReadyQueue` similarly updated.
- `specs/ticket_snooze.feature`: Two new scenarios added for future-snoozed exclusion and past-snooze inclusion.
- `testdata/ticket_snooze/ready_exclusion.txtar`: Integration test with `ko-future` (snooze: 2099-01-01) and `ko-past` (snooze: 2020-01-01); confirmed both behaviors with `ko ready`.

## Plan compliance

All 6 planned tasks completed as specified. No deviations.

## Notable decisions

- The `IsSnoozed` boundary: `now.Before(parsedDate)` (strictly less than midnight) means a ticket snoozed until "2026-03-15" becomes ready at exactly `2026-03-15T00:00:00Z`, matching the ticket's "valid as of midnight" specification.
- The txtar test runs `exec ko ready` twice (once to check absence of `ko-future`, once to confirm presence of `ko-past`). Functionally correct since both assertions apply to the same steady-state ticket set; could have been a single exec with two assertions but poses no issue.

## Test results

`go test ./... -run "TestIsSnoozed|TestTicketSnooze"` — passed.
`go build ./...` — passed.
