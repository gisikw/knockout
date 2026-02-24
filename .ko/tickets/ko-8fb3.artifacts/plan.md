## Goal
Add structured JSONL logging to agent runs, recording run-level metrics (tickets processed, outcomes, duration, stop reason) appended to `.ko/agent.log`.

## Context
The codebase already has extensive JSONL logging infrastructure:

- `eventlog.go` — EventLogger writes per-ticket build events to a file specified by `KO_EVENT_LOG` env var. Already has `LoopSummary()` method that emits a `loop_summary` event with processed counts and stop reason.
- `.ko/tickets/<id>.jsonl` — Per-ticket build events (workflow/node start/complete) are written here via build history mechanism.
- `.ko/agent.log` — Currently text output from agent loop runs (lines like "loop: building ko-xxx — title", "loop: ko-xxx SUCCEED", "loop complete: N processed...").
- `loop.go:RunLoop()` — Returns `LoopResult` struct with processed counts and stop reason. Already calls `log.LoopSummary(result)` at the end.
- `cmd_loop.go:cmdAgentLoop()` — Creates EventLogger with `OpenEventLog()`, runs loop, calls `log.LoopSummary()`, prints text summary.
- `cmd_agent.go:cmdAgentStart()` — Daemonizes loop, redirects stdout/stderr to `.ko/agent.log`.

Current state:
- Per-ticket build events go to `.ko/tickets/<id>.jsonl` (already implemented in ko-bfc5).
- Loop summaries already emit to `KO_EVENT_LOG` when set, but `.ko/agent.log` only gets text output.
- The ticket suggests "Maybe .ko/agent.log jsonl?" — implying agent.log should contain JSONL run summaries.

Project conventions (from INVARIANTS.md and tests):
- Pure decision functions separate from I/O.
- Event emission is append-only JSONL with timestamps.
- No multi-purpose functions.

## Approach
Modify `cmdAgentLoop()` to write a JSONL summary line to `.ko/agent.log` after each loop run completes. Keep the existing text output for human readability (prepend text summary lines with "loop: " or similar for easy filtering), then append one JSONL line with structured run metadata.

The JSONL summary will include: timestamp, tickets processed, outcome counts (succeeded/failed/blocked/decomposed), stop reason, and total runtime.

## Tasks
1. [cmd_loop.go:cmdAgentLoop] — After the loop completes and summary is printed, append a single JSONL line to `.ko/agent.log` with structured run summary (tickets_processed, succeeded, failed, blocked, decomposed, stop_reason, runtime_seconds, timestamp).
   Verify: Run `ko agent loop` and check that `.ko/agent.log` contains both text output and a final JSONL line with the run summary.

2. [cmd_loop.go:cmdAgentLoop] — Calculate and include runtime duration in the JSONL summary by capturing start time before `RunLoop()` and computing elapsed time after.
   Verify: JSONL line includes `runtime_seconds` field with correct elapsed time.

3. [loop_test.go] — Add a test that verifies LoopResult contains the expected fields used for the summary.
   Verify: `go test ./... -run TestLoopResult` passes.

## Open Questions
None. The implementation is straightforward: append JSONL to the existing `.ko/agent.log` file alongside human-readable text. This mirrors the pattern where `.ko/tickets/<id>.jsonl` contains structured events while workspace files contain markdown output.

The ticket mentions "if we don't have this yet" — we have partial implementation (per-ticket JSONL logs and EventLogger with LoopSummary), but not run-level summaries in agent.log itself. This plan completes that gap.
