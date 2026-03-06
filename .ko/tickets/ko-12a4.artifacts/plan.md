## Goal
Route `ko agent start` and `ko agent stop` success messages to stderr so that stdout remains clean for composable use.

## Context
- `cmdAgentStart` (`cmd_agent.go:178`) prints its success message with `fmt.Printf`, which writes to stdout. All other messages in that function already use `fmt.Fprintf(os.Stderr, ...)`.
- `cmdAgentStop` (`cmd_agent.go:227`) has the same stdout pollution pattern: `fmt.Printf("agent stopped (pid %d)\n", pid)`.
- `maybeAutoAgent` (mentioned in prior plan notes) does not exist in the codebase yet. The prior plan's Task 2 about updating a comment in that function is moot.
- Author has confirmed: fix both start and stop in this ticket, and stderr for direct invocations is acceptable (standard CLI behavior).

## Approach
Two one-line changes: replace `fmt.Printf` with `fmt.Fprintf(os.Stderr, ...)` at lines 178 and 227. No other changes needed.

## Tasks
1. [`cmd_agent.go:178`] Change `fmt.Printf("agent started (pid %d), logging to %s\n", cmd.Process.Pid, logPath)` to `fmt.Fprintf(os.Stderr, "agent started (pid %d), logging to %s\n", cmd.Process.Pid, logPath)`.
   Verify: `go build ./...` passes.

2. [`cmd_agent.go:227`] Change `fmt.Printf("agent stopped (pid %d)\n", pid)` to `fmt.Fprintf(os.Stderr, "agent stopped (pid %d)\n", pid)`.
   Verify: `go build ./...` passes.

## Open Questions
None.
