The plan is written. Here's a summary:

## Plan Summary

**Goal:** When the agent loop runs (via `ko agent start` or `ko agent loop`), it triages all tickets with a `triage` field set before processing any ready tickets.

**Key findings:**
- `ko agent start` is just a daemon launcher for `ko agent loop` — the real change goes into the loop
- The core triage execution in `cmdAgentTriage` is currently inlined (not reusable)
- There's an existing inconsistency: `ReadyQueue()` doesn't exclude triaged tickets, but `cmdReady` does — this needs fixing too

**6 tasks:**
1. **`cmd_agent_triage.go`** — Extract `runAgentTriage(ticketsDir, ticket, pipeline, verbose) error` as a shared function; CLI becomes a thin wrapper
2. **`loop.go:ReadyQueue`** — Add `&& t.Triage == ""` filter (safety net so failed-triage tickets don't get built)
3. **`loop.go`** — Add `TriageQueue()` (pure query) + `runTriagePass()` (orchestrator), and call `runTriagePass()` at the top of each `RunLoop` iteration
4. **`loop_test.go`** — Unit test for `TriageQueue()`
5. **`specs/loop.feature`** — Two new scenarios for the triage pre-pass behavior
6. **`testdata/loop/loop_triage_before_ready.txtar`** — Integration testscript

**3 open questions** for the implementer:
1. Should triage failures log-and-continue or stop the loop?
2. Should triaged tickets count toward `--max-tickets`?
3. Should stop signals interrupt mid-triage or wait for the next gap?
