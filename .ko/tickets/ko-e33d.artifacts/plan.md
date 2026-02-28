## Goal
When `pipeline.yml` (or `config.yaml`) has `auto_triage: true`, creating or updating a ticket with a non-empty `triage` field automatically runs `ko agent triage` against it.

## Context

**Pipeline config** (`pipeline.go`): `Pipeline` struct holds all pipeline settings. New fields are parsed in `ParsePipeline` (top-level scalars block, ~line 200–235). `pipeline.go` is already at 609 lines (over the 500-line invariant), so additions must be minimal.

**`ko agent triage`** (`cmd_agent_triage.go:cmdAgentTriage`): Loads pipeline, constructs a prompt from ticket content + triage instructions, runs the agent, then clears the `triage` field on success. Currently structured as one monolithic CLI entrypoint — does not expose a callable function for internal use.

**Ticket creation** (`cmd_create.go:cmdCreate`): Saves the ticket then emits a mutation event. The `--triage` flag sets `t.Triage` before save (line 181–183).

**Ticket update** (`cmd_update.go:cmdUpdate`): Same pattern — `--triage` sets `t.Triage` (line 239–242), saved before the mutation event.

**`ko triage <id> <instructions>`** (`cmd_list.go:cmdTriage`, line 313–314): Calls `cmdUpdate` with `--triage=<instructions>`, so auto-triage via that path is handled by covering `cmdUpdate`.

**Testing pattern**: Behavior specs in `specs/*.feature`, integration tests in `testdata/<domain>/*.txtar` via testscript. New behaviors need both a spec scenario and a txtar test.

**File sizes**: `cmd_create.go` (372 lines), `cmd_update.go` (262 lines), `cmd_agent_triage.go` (161 lines) — all well within the 500-line limit.

## Approach

Add `AutoTriage bool` to `Pipeline` and parse `auto_triage: true` in `ParsePipeline`. Extract a `runAgentTriage(ticketsDir, id string, verbose bool) int` helper from `cmdAgentTriage`, and add a `maybeAutoTriage(ticketsDir, id string)` function that silently skips if no config exists or `AutoTriage` is false, and warns (but does not fail) if the triage run itself fails. Call `maybeAutoTriage` from both `cmdCreate` and `cmdUpdate` whenever a non-empty triage value is being saved.

Auto-triage failure is non-fatal: the ticket is already saved with the triage field set, the user sees a stderr warning, and can manually run `ko agent triage` afterward.

## Tasks

1. **[`pipeline.go:Pipeline`]** — Add `AutoTriage bool` field to the `Pipeline` struct. In `ParsePipeline`, in the top-level scalar `switch` block, add a `case "auto_triage": p.AutoTriage = val == "true"` branch. Similarly handle it in `ParseConfig`'s pipeline passthrough (no change needed there — it strips 2-space indent and feeds to `ParsePipeline`).
   Verify: `TestParsePipeline*` tests still pass; add `TestParsePipelineAutoTriage` in `pipeline_test.go` confirming `auto_triage: true` sets `AutoTriage = true` and absence defaults to `false`.

2. **[`cmd_agent_triage.go`]** — Refactor `cmdAgentTriage` to extract a `runAgentTriage(ticketsDir, id string, verbose bool) int` function containing all logic from the ticket-load step onward (line ~47 to end). `cmdAgentTriage` becomes a thin wrapper: resolve ticketsDir from CLI args, resolve the ticket ID, then delegate to `runAgentTriage`. Add a `maybeAutoTriage(ticketsDir, id string)` function that: (a) tries `FindConfig`; if error, returns silently; (b) loads the pipeline config; (c) checks `p.AutoTriage`; if false, returns; (d) calls `runAgentTriage(ticketsDir, id, false)`; if it returns non-zero, prints a stderr warning like `"ko: auto-triage for <id> failed; run 'ko agent triage <id>' manually"` and returns.
   Verify: existing `testdata/agent_triage/*.txtar` tests still pass.

3. **[`cmd_create.go:cmdCreate`]** — After `EmitMutationEvent` (line 190), if `*triage != ""`, call `maybeAutoTriage(ticketsDir, t.ID)`. The function is non-blocking and handles its own error output; `cmdCreate` continues to `fmt.Println(t.ID)` and return 0 regardless.
   Verify: `testdata/ticket_creation/` tests still pass; new txtar added in step 5.

4. **[`cmd_update.go:cmdUpdate`]** — After `EmitMutationEvent` (line 256), if `*triage != ""`, call `maybeAutoTriage(ticketsDir, id)`. Same non-blocking pattern.
   Verify: `testdata/ticket_triage/` tests still pass; new txtar added in step 5.

5. **[`specs/ticket_triage.feature`]** — Add scenarios:
   - *`auto_triage: true` triggers triage after `ko add --triage`*: Given a pipeline with `auto_triage: true` and a mock harness, `ko add 'Task' --triage 'unblock this'` should create the ticket and automatically clear its triage field.
   - *`auto_triage: true` triggers triage after `ko update --triage`*: Same pipeline; `ko update <id> --triage 'break apart'` should clear triage automatically.
   - *`auto_triage: false` (or absent) does not trigger*: With no `auto_triage` key, `ko add 'Task' --triage 'unblock'` leaves the triage field set.
   - *Auto-triage failure is non-fatal*: A harness that exits 1 makes `ko add` still succeed with exit 0, but the triage field remains set and a warning appears on stderr.
   Verify: scenarios match txtar tests.

6. **[`testdata/ticket_triage/`]** — Add txtar tests covering the four new scenarios above: `auto_triage_on_create.txtar`, `auto_triage_on_update.txtar`, `auto_triage_absent_no_trigger.txtar`, `auto_triage_failure_nonfatal.txtar`. Use the same mock harness pattern as `testdata/agent_triage/triage_run.txtar`.
   Verify: `go test ./...` passes with `testscript`.

## Open Questions

1. **Should auto-triage failure be fatal for `ko add`/`ko update`?** Currently planned as non-fatal (warning only). The ticket was already created/saved; leaving it in the ready-for-manual-triage state seems safer than rolling back. Confirm this is the desired behavior.

2. **`pipeline.go` file size**: The file is at 609 lines, already over the 500-line invariant. Adding ~5 lines for `AutoTriage` is minor but technically makes it worse. A split of `pipeline.go` is out of scope for this ticket; proceed with the minimal addition and note the debt.
