Done. Here's a summary of changes:

**Task 1 — `cmd_status.go`:** Added `cmdSnooze(args []string) int` between `cmdOpen` and `cmdBlock`. Requires 2 args; delegates to `cmdUpdate([]string{args[0], "--snooze=" + args[1]})`.

**Task 2 — `main.go:run()`:** Added `case "snooze": return cmdSnooze(rest)` after the `block` case.

**Task 3 — `main.go:cmdHelp()`:** Added `snooze <id> <date>  Snooze ticket until date (ISO 8601, e.g. 2026-05-01)` alongside the other status shortcuts.

**Task 4 — `cmd_status_test.go`:** Added `TestCmdSnooze` with three cases: valid date sets snooze, missing date arg returns error, invalid date format returns error.

No deviations from the plan.
