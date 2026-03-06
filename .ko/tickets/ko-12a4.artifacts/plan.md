## Goal
Route the "agent started" confirmation message to stderr so that `ko create` (and other mutation commands with auto-start) remain composable via stdout.

## Context
- `cmdAgentStart` (`cmd_agent.go:178`) prints its success message with `fmt.Printf`, which writes to stdout.
- `maybeAutoAgent` (`cmd_agent.go:372–401`, uncommitted) invokes `ko agent start` as a subprocess and works around the stdout pollution by setting `cmd.Stdout = os.Stderr`. This is a functional workaround but the root cause is the `fmt.Printf` in `cmdAgentStart`.
- All other user-facing messages in `cmdAgentStart` (errors, warnings) already use `fmt.Fprintf(os.Stderr, ...)` — the success message at line 178 is the only outlier.
- `maybeAutoAgent` is called from `cmdCreate` (after printing the ticket ID), `cmdUpdate`, and `cmdDep` (both add and remove).
- There are no existing tests for `maybeAutoAgent` or `cmdAgentStart`'s output streams.

## Approach
Change the one `fmt.Printf` in `cmdAgentStart` to `fmt.Fprintf(os.Stderr, ...)`. This makes the fix canonical: the message goes to stderr regardless of how `ko agent start` is invoked (directly or via auto-start). The `cmd.Stdout = os.Stderr` line in `maybeAutoAgent` becomes redundant but is harmless; update its comment to reflect that it's now a safety net rather than the primary fix.

## Tasks
1. [`cmd_agent.go:178`] Change `fmt.Printf("agent started (pid %d), logging to %s\n", ...)` to `fmt.Fprintf(os.Stderr, "agent started (pid %d), logging to %s\n", ...)`.
   Verify: `go build ./...` passes. Running `ko agent start` manually still shows the message (now on stderr).

2. [`cmd_agent.go:399`] Update the comment on `cmd.Stdout = os.Stderr` from "surface start message without polluting stdout" to something like "safety net: agent start no longer writes to stdout, but keep stderr routing in case of future regressions". Or simplify to `cmd.Stdout = nil` since the message now goes to stderr natively and there's nothing useful to capture from stdout.
   Verify: `go build ./...` passes.

3. [No test file change required] There are no existing stream-capture tests for `cmdAgentStart`. Since the fix is a single-line `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` change with no logic involved, no new test is strictly needed. If the project adds integration tests later, this behavior is verifiable by capturing stdout of `ko create` and asserting the agent message is absent.

## Open Questions
- **`ko agent stop` consistency**: Line 227 has `fmt.Printf("agent stopped (pid %d)\n", ...)` — same stdout pattern. The ticket only mentions agent start, but fixing stop to use stderr too would be consistent. Leaving it for a follow-up unless the author wants it included here.
- **Direct invocation UX**: Changing line 178 to stderr means `ko agent start` run interactively also prints to stderr. This is standard CLI behavior (status messages → stderr, data → stdout) and shouldn't surprise users, but worth confirming the author agrees.
