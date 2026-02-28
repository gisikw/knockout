All tasks are complete. All 6 plan tasks were already implemented from a previous build attempt. Here's a summary of what's in place:

**Task 1 — `pipeline.go`**: `AutoTriage bool` field added to `Pipeline` struct; `case "auto_triage": p.AutoTriage = val == "true"` in `ParsePipeline`.

**Task 1 (tests) — `pipeline_test.go`**: `TestParsePipelineAutoTriage` with three sub-tests: `true`, `false`, and absent (defaults to false).

**Task 2 — `cmd_agent_triage.go`**: `cmdAgentTriage` is a thin wrapper that delegates to `runAgentTriage(ticketsDir, id string, verbose bool) int`. `maybeAutoTriage(ticketsDir, id string)` silently skips if no config or `AutoTriage=false`; prints a non-fatal stderr warning if triage fails.

**Task 3 — `cmd_create.go`**: After `EmitMutationEvent`, calls `maybeAutoTriage(ticketsDir, t.ID)` when `*triage != ""`.

**Task 4 — `cmd_update.go`**: Same pattern after its `EmitMutationEvent`.

**Task 5 — `specs/ticket_triage.feature`**: Four new scenarios covering auto-trigger on create, auto-trigger on update, no-trigger when absent, and non-fatal failure.

**Task 6 — `testdata/ticket_triage/`**: Four new txtar tests with mock harnesses (`auto_triage_on_create.txtar`, `auto_triage_on_update.txtar`, `auto_triage_absent_no_trigger.txtar`, `auto_triage_failure_nonfatal.txtar`).

`go test -count=1 ./...` passes in 10.4s. No deviations from the plan.
