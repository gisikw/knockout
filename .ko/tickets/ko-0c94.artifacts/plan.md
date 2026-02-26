## Goal
Remove the `ko triage` command and the `ko blocked` command, which are superseded by `ko update`.

## Context
The codebase has a complete triage command implementation in `cmd_triage.go` with extensive test coverage in `cmd_triage_test.go`. The command provides:
- Bare invocation to show triage state (block reason and questions)
- `--block [reason]` flag to block a ticket
- `--questions <json>` flag to add questions and implicitly block
- `--answers <json>` flag to answer questions and auto-unblock when done
- `--json` flag for JSON output

The `ko blocked` command is defined in `cmd_list.go:cmdBlocked` and shows blocked tickets or a specific ticket's block reason.

Both commands are registered in `main.go:run()` in the switch statement (lines 38 and 45).

Key helper functions used by triage:
- `ExtractBlockReason(t *Ticket) string` - defined in `ticket.go`, used by both `cmdTriage` and `cmdBlocked`, also used by tests
- `ValidatePlanQuestions([]PlanQuestion) error` - defined in `cmd_status.go`, used by `cmdTriage`, `cmdUpdate`, and `disposition.go`

The ticket states that `ko update` supersedes `ko triage`. Checking `cmd_update.go` confirms it has `--questions` and `--answers` flags that provide the same functionality.

## Approach
1. Remove `cmd_triage.go` and `cmd_triage_test.go` files entirely
2. Remove the `cmdBlocked` function from `cmd_list.go`
3. Remove the switch cases for `"triage"` and `"blocked"` from `main.go`
4. Remove references to triage and blocked from the help text in `main.go:cmdHelp`
5. Keep `ExtractBlockReason` and `ValidatePlanQuestions` as they are used by other commands
6. Run tests to ensure nothing else breaks

The removal is clean because:
- `ExtractBlockReason` is still used by the remaining `cmdBlocked` display logic (wait, we're removing cmdBlocked too) - actually it's only used by cmdTriage and cmdBlocked, but since we're removing both, we should evaluate if it's still needed
- Actually, checking again: `ExtractBlockReason` is used in `cmd_list.go:cmdBlocked` (lines 339, 410) which we're removing, and in `cmd_triage.go` which we're removing. So `ExtractBlockReason` should also be removed.
- `ValidatePlanQuestions` is used by `cmd_update.go` and `disposition.go`, so it must be kept

## Tasks
1. [cmd_triage.go] — Delete entire file.
   Verify: File no longer exists.

2. [cmd_triage_test.go] — Delete entire file.
   Verify: File no longer exists.

3. [cmd_list.go:cmdBlocked] — Delete the entire `cmdBlocked` function (lines 309-427).
   Verify: Function is removed, file compiles.

4. [main.go:run] — Remove the switch case `case "triage":` at line 38-39.
   Verify: Switch statement no longer has triage case.

5. [main.go:run] — Remove the switch case `case "blocked":` at line 44-45.
   Verify: Switch statement no longer has blocked case.

6. [main.go:cmdHelp] — Remove triage and blocked commands from help text (lines 134-137 mention triage, line 125 mentions blocked).
   Verify: Help text no longer references triage or blocked commands.

7. [ticket.go:ExtractBlockReason] — Delete the `ExtractBlockReason` function (lines 625-664) and its comment since it's only used by the commands we're removing.
   Verify: Function is removed, file compiles.

8. [ticket_test.go:TestExtractBlockReason] — Delete the `TestExtractBlockReason` test function since we're removing the function it tests.
   Verify: Test is removed, tests compile.

9. Run `go test ./...` to verify all remaining tests pass.
   Verify: All tests pass with no failures.

10. Run `go build` to ensure the project compiles.
    Verify: Build succeeds with no errors.

## Open Questions
None. This is a straightforward deletion task. The `ko update` command already provides the functionality of `ko triage` via its `--questions` and `--answers` flags, and the `ko blocked` command is no longer needed. All helper functions are either removed with their only callers or retained because they're used by other commands.
