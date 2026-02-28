All done. Here's a summary of what was implemented:

**`pipeline.go`** — Added `AutoTriage bool` field to `Pipeline` struct and a `case "auto_triage": p.AutoTriage = val == "true"` branch in `ParsePipeline`.

**`pipeline_test.go`** — Added `TestParsePipelineAutoTriage` with three sub-tests covering `true`, `false`, and absent (defaults to false).

**`cmd_agent_triage.go`** — Extracted `runAgentTriage(ticketsDir, id string, verbose bool) int` from `cmdAgentTriage` (which became a thin wrapper). Added `maybeAutoTriage(ticketsDir, id string)` which silently skips if no config or `AutoTriage=false`, and prints a non-fatal warning if the triage run fails.

**`cmd_create.go`** — After `EmitMutationEvent`, calls `maybeAutoTriage` if `--triage` was set.

**`cmd_update.go`** — Same pattern after its `EmitMutationEvent`.

**`specs/ticket_triage.feature`** — Added 4 new scenarios covering auto-trigger on create, auto-trigger on update, no-trigger when absent, and non-fatal failure.

**`testdata/ticket_triage/`** — Added 4 txtar integration tests with mock harnesses verifying each scenario.
