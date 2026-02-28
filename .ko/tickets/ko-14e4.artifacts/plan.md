## Goal
When `ko agent loop` runs (including when started via `ko agent start`), it first runs `ko agent triage` on every ticket that has a `triage` field set before processing any ready tickets.

## Context

**`ko agent start`** (`cmd_agent.go:cmdAgentStart`) is a thin daemon launcher: it re-execs itself as `ko agent loop` with `Setsid`. The actual work happens in the loop.

**`ko agent loop`** (`cmd_loop.go:cmdAgentLoop`) calls `RunLoop()` (`loop.go:RunLoop`), which processes the ready queue one ticket at a time in a `for` loop.

**`ko agent triage <id>`** (`cmd_agent_triage.go:cmdAgentTriage`) runs the AI adapter against a ticket whose `triage` field is non-empty, then clears the `triage` field on success. All execution logic is currently inline in `cmdAgentTriage`; there is no shared helper.

**`ReadyQueue()`** (`loop.go:ReadyQueue`) does NOT exclude tickets with `triage != ""`, even though `cmdReady` in `cmd_list.go:230` does. This means a ready ticket with a pending triage could be built without being triaged first — a bug that the new triage pre-pass would expose if not also fixed.

**Invariants to respect:**
- Decision logic is pure (new `TriageQueue` should be a pure query)
- No multi-purpose functions (separate query from execution)
- Each file stays under 500 lines (`cmd_agent_triage.go` is 161 lines, `loop.go` is 191 lines — both have room)
- Spec before code — new behavior needs a spec + txtar test

## Approach

Extract the core execution logic from `cmdAgentTriage` into a shared `runAgentTriage(ticketsDir string, t *Ticket, p *Pipeline, verbose bool) error` function. Add a `TriageQueue()` query and a `runTriagePass()` orchestrator in `loop.go`. Call `runTriagePass()` at the top of each `RunLoop` iteration (after signal/limits checks, before `ReadyQueue()`). Also fix `ReadyQueue()` to exclude tickets with `triage != ""` so that a ticket the triage pass fails on is never accidentally built.

## Tasks

1. **[`cmd_agent_triage.go`]** — Extract the core execution logic (prompt construction, artifact dir creation, adapter invocation, triage-clear-on-success) from `cmdAgentTriage` into a new package-level function:
   ```
   func runAgentTriage(ticketsDir string, t *Ticket, p *Pipeline, verbose bool) error
   ```
   `cmdAgentTriage` becomes a thin wrapper: resolve args, load ticket, check `t.Triage != ""`, load pipeline, call `runAgentTriage`, print `"%s: triage cleared"`.
   Verify: `go test ./...` passes (existing `triage_run.txtar` still passes).

2. **[`loop.go:ReadyQueue`]** — Add `&& t.Triage == ""` to the filter condition in `ReadyQueue()`, matching what `cmdReady` already does. A ticket pending triage is not ready to build.
   Verify: existing loop txtar tests pass; `loop_succeed.txtar` (tickets have no triage) still reports `2 processed`.

3. **[`loop.go`]** — Add two functions:
   - `TriageQueue(ticketsDir string) ([]*Ticket, error)` — returns all tickets with `t.Triage != ""`, sorted by priority then modified. Pure query.
   - `runTriagePass(ticketsDir string, p *Pipeline, verbose bool, stop <-chan struct{}) int` — iterates `TriageQueue()`, logs each ticket being triaged, calls `runAgentTriage` per ticket, logs failures without stopping the loop, respects `stop` channel between runs. Returns count of tickets triaged.

   In `RunLoop`, after the stop-signal check and limits check (but before `ReadyQueue()`), add:
   ```
   runTriagePass(ticketsDir, p, config.Verbose, stop)
   ```
   Log progress with `fmt.Printf("loop: triaging %s — %s\n", id, t.Title)` (only when `!config.Quiet`).
   Verify: `go test ./...` passes.

4. **[`loop_test.go`]** — Add `TestTriageQueue` (unit test for the new pure query function): creates temp tickets with and without `triage` set, asserts that `TriageQueue` returns only the triaged ones.
   Verify: new test passes.

5. **[`specs/loop.feature`]** — Add two scenarios under a new `# Triage pre-pass` section:
   - "Loop triages all triageable tickets before processing ready tickets" — given one ticket with `triage` set and one without, when loop runs, triage runs first, then the ready ticket is built.
   - "Loop continues building ready tickets even if triage fails for one ticket" — given a ticket whose triage fails (mock harness exits non-zero), the loop still processes other ready tickets.
   Verify: reviewed for correctness against the implementation.

6. **[`testdata/loop/loop_triage_before_ready.txtar`]** — Add a testscript that has:
   - `ko-a001`: status `open`, `triage: unblock this ticket` (so it appears in triage queue but not ready queue until triage clears the field)
   - `ko-b002`: status `open`, no triage (should be built after triage of a001 completes)
   - A `fake-llm` script that succeeds for both triage and build
   - Assert: loop output contains "triaging ko-a001", both tickets end up resolved
   Verify: `go test -run TestScript/loop/loop_triage_before_ready ./...` passes.

## Open Questions

1. **Triage failure behavior**: If `runAgentTriage` fails for a ticket (agent exits non-zero, timeout, etc.), should the loop log and continue to the next ticket — or stop the loop with `build_error`? The plan assumes log-and-continue (consistent with how FAIL outcomes don't stop the loop). Confirm this is the desired behavior.

2. **Triage counting toward MaxTickets**: Should triaged tickets count toward the `--max-tickets` limit? The plan assumes NO — triage is a pre-build housekeeping step, not a ticket build. Confirm.

3. **Stop signal during triage**: Should a stop signal received while triage is running for a ticket be honored immediately (interrupt current triage) or at the next inter-ticket gap? The plan assumes inter-ticket only (check `stop` between each triage run, not mid-run), consistent with how stop is checked between builds today.
