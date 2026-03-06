Plan written to `.ko/tickets/ko-12a4.artifacts/plan.md`. Here's the summary:

**Root cause**: `cmdAgentStart` uses `fmt.Printf` (stdout) for its success message at line 178 while every other message in that function already uses `fmt.Fprintf(os.Stderr, ...)`.

**Fix**: One-line change — `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` at `cmd_agent.go:178`. The `cmd.Stdout = os.Stderr` workaround already in `maybeAutoAgent` becomes redundant but harmless.

**Open questions flagged**:
1. Should `ko agent stop` (line 227, same pattern) be fixed for consistency in this ticket?
2. Does the author agree with "agent start status message goes to stderr even on direct invocation" — which is idiomatic but worth confirming?
