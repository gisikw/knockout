Plan written to `.ko/tickets/ko-1d18.artifacts/plan.md`. Here's the summary:

**Goal:** Add `ko snooze <id> <date>` as a shorthand that delegates to `ko update <id> --snooze=<date>`.

**Approach:** Follow the exact pattern used by `cmdStart`/`cmdClose`/`cmdOpen` in `cmd_status.go` — a thin function that validates arg count and calls `cmdUpdate` with the appropriate flag.

**4 tasks:**
1. **`cmd_status.go`** — Add `cmdSnooze` function (2-arg check, delegate to `cmdUpdate` with `--snooze=<date>`)
2. **`main.go:run()`** — Add `case "snooze"` to the switch
3. **`main.go:cmdHelp()`** — Add snooze to help text
4. **`cmd_status_test.go`** — Add `TestCmdSnooze` covering valid date, missing date, and invalid date format

No open questions — the pattern is fully established by existing shorthand commands.
