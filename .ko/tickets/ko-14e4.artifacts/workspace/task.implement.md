All six tasks complete. Here's a summary:

---

## Summary

### Task 1 — `cmd_agent_triage.go`
Refactored `runAgentTriage` signature from `(ticketsDir, id string, verbose bool) int` to `(ticketsDir string, t *Ticket, p *Pipeline, verbose bool) error`. The function now takes pre-loaded ticket and pipeline and returns an error instead of an exit code. `cmdAgentTriage` is the thin wrapper that resolves the ticket ID, loads the ticket and pipeline, calls `runAgentTriage`, and prints `"%s: triage cleared"`. Updated `maybeAutoTriage` to load the ticket and pipeline before calling the new `runAgentTriage`.

### Task 2 — `loop.go:ReadyQueue`
Added `&& t.Triage == ""` to the filter condition, matching the existing behavior in `cmdReady`. Tickets with a pending triage field are no longer included in the build queue.

### Task 3 — `loop.go`
Added `TriageQueue(ticketsDir string) ([]*Ticket, error)` (pure query returning tickets with `triage != ""`). Added `runTriagePass(ticketsDir string, p *Pipeline, verbose bool, quiet bool, stop <-chan struct{}) int` (orchestrates triage, logs failures without stopping, respects stop signal between runs). Called `runTriagePass` at the top of each `RunLoop` iteration, before `ReadyQueue`.

**Minor deviation from plan**: Added `quiet bool` to `runTriagePass` since the plan text says to log only when `!config.Quiet` but the written signature omitted it. This is clearly the intended behavior.

### Task 4 — `loop_test.go`
Added `TestTriageQueue` unit test: creates two temp tickets (one with `triage` set, one without), asserts that `TriageQueue` returns only the triaged one.

### Task 5 — `specs/loop.feature`
Added two scenarios under a new `# Triage pre-pass` section: one for successful triage before build, one for triage failure not blocking ready tickets.

### Task 6 — `testdata/loop/loop_triage_before_ready.txtar`
Added txtar with `ko-a001` (triage set), `ko-b002` (no triage), and a `fake-llm` that always succeeds. Asserts loop output contains `"loop: triaging ko-a001"` and both tickets end up resolved.
