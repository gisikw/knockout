The plan is written. Here's a summary:

## Plan: Auto-Triage on Ticket Create/Update

**Goal**: When `auto_triage: true` is set in the pipeline config, creating or updating a ticket with a non-empty `triage` field automatically runs `ko agent triage`.

**6 tasks:**

1. **`pipeline.go`** — Add `AutoTriage bool` to `Pipeline` struct; parse `auto_triage: true/false` in `ParsePipeline`.

2. **`cmd_agent_triage.go`** — Extract `runAgentTriage(ticketsDir, id, verbose)` from `cmdAgentTriage` (so it can be called programmatically), and add `maybeAutoTriage(ticketsDir, id)` which silently skips if no config or `AutoTriage=false`, and warns-but-doesn't-fail if the triage run fails.

3. **`cmd_create.go`** — After `EmitMutationEvent`, if `--triage` was set, call `maybeAutoTriage`.

4. **`cmd_update.go`** — Same, after saving the ticket. (This also covers `ko triage <id> <instructions>`, which routes through `cmdUpdate`.)

5. **`specs/ticket_triage.feature`** — Add 4 new scenarios covering auto-trigger, no-trigger, and non-fatal failure.

6. **`testdata/ticket_triage/`** — Add 4 txtar integration tests with mock harnesses.

**Two open questions flagged**: (1) whether auto-triage failure should be fatal for `ko add`/`ko update` (planned as non-fatal), and (2) `pipeline.go` already exceeds the 500-line invariant.
