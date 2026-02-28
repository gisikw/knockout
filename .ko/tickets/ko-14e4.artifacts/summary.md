# Summary: ko-14e4 — Triage pre-pass in agent loop

## What Was Done

The implementation added a triage pre-pass to the agent loop so that every iteration triages all tickets with a `triage` field set before processing any ready tickets.

**Changes:**

1. **`cmd_agent_triage.go`** — Extracted core triage execution into `runAgentTriage(ticketsDir string, t *Ticket, p *Pipeline, verbose bool) error`. The CLI command and `maybeAutoTriage` became thin wrappers. `maybeAutoTriage` now loads the ticket and pipeline itself before calling the shared function.

2. **`loop.go`** — Added `TriageQueue()` (pure query returning tickets with `triage != ""`), `runTriagePass()` (iterates queue, logs failures, respects stop channel), and wired `runTriagePass()` at the top of each `RunLoop` iteration (after stop/limits checks, before `ReadyQueue()`). Also fixed `ReadyQueue()` to exclude tickets with `triage != ""` — a pre-existing inconsistency that would have allowed untriaged tickets to be built.

3. **`loop_test.go`** — Added `TestTriageQueue` unit test.

4. **`specs/loop.feature`** — Added two scenarios under `# Triage pre-pass`.

5. **`testdata/loop/loop_triage_before_ready.txtar`** — Integration test for the success path.

## Notable Decisions

- **`runTriagePass` signature includes `quiet bool`** — The plan's function signature omitted this, but the description required `!config.Quiet` gating on the progress log. Adding `quiet` was necessary and self-evident.

- **`maybeAutoTriage` required re-expansion** — When `runAgentTriage` changed from taking `(ticketsDir, id string, verbose bool)` to `(ticketsDir string, t *Ticket, p *Pipeline, verbose bool)`, `maybeAutoTriage` had to load the ticket and pipeline itself. The refactored version also gained an early exit when `t.Triage == ""`, which is cleaner than the original.

## Invariant Fix Applied During Review

The second spec scenario ("Loop continues building ready tickets even if triage fails") lacked a corresponding txtar test — an INVARIANTS violation ("A spec without a corresponding test is an unverified claim"). Added `testdata/loop/loop_triage_fail_continues.txtar` to cover it.

The fake-llm in this test distinguishes triage from build calls using `KO_BUILD_HISTORY`: build stages always set this env var; `runAgentTriage` does not. This is a reliable structural difference, not a content-based heuristic.

## All Tests Pass

`go test ./...` passes including all new tests.
