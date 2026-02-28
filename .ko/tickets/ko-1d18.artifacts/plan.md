## Goal
Add `ko snooze <id> <date>` as a shorthand command that delegates to `ko update <id> --snooze=<date>`.

## Context
- `cmd_status.go` contains `cmdStart`, `cmdClose`, `cmdOpen`, `cmdBlock`, and `cmdStatus` — all shorthand commands that delegate to `cmdUpdate` by constructing the appropriate flag string and calling `cmdUpdate(args)`. `cmdSnooze` follows this exact pattern.
- `cmd_update.go` already handles `--snooze` with ISO 8601 date validation (`time.Parse("2006-01-02", ...)`).
- `main.go` contains the `run()` switch statement and `cmdHelp()` — both need updating.
- `cmd_status_test.go` and `cmd_update_test.go` show the test patterns: table-driven tests using `t.TempDir()`, saving a ticket, changing dir, suppressing stdout/stderr, calling the cmd function, and loading+asserting the result.

## Approach
Add `cmdSnooze` to `cmd_status.go` (it's the natural home for shorthand status/field commands). Wire it into the switch in `main.go`. Add it to the help text. Add tests to `cmd_status_test.go`.

## Tasks
1. [cmd_status.go] — Add `cmdSnooze(args []string) int`. Require at least 2 args (`<id>` and `<date>`); print usage error and return 1 if not. Delegate to `cmdUpdate([]string{args[0], "--snooze=" + args[1]})`.
   Verify: `go build ./...` succeeds.

2. [main.go:run()] — Add `case "snooze": return cmdSnooze(rest)` to the switch statement.
   Verify: `ko snooze` appears as a valid command (no "unknown command" error).

3. [main.go:cmdHelp()] — Add `snooze <id> <date>     Snooze ticket until date (ISO 8601, e.g. 2026-05-01)` to the help text, near the `update` entry or alongside other status shortcuts.
   Verify: `ko help` output includes the snooze line.

4. [cmd_status_test.go] — Add `TestCmdSnooze` covering: (a) valid date sets snooze field, (b) missing date arg returns error, (c) invalid date format returns error.
   Verify: `go test ./... -run TestCmdSnooze` passes.

## Open Questions
None. The implementation is fully constrained by the existing pattern.
