# After-Action Summary: ko-e33d — Auto-Triage on Ticket Create/Update

## What Was Done

Implemented automatic triggering of `ko agent triage` when a ticket is created
or updated with a `triage` field set, conditional on `auto_triage: true` in
`pipeline.yml`.

### Changes

- **`pipeline.go`** — Added `AutoTriage bool` to `Pipeline` struct; added
  `case "auto_triage"` in `ParsePipeline`. (+4 lines)

- **`pipeline_test.go`** — `TestParsePipelineAutoTriage` with three sub-tests
  covering `true`, `false`, and absent (defaults to `false`).

- **`cmd_agent_triage.go`** — Extracted `runAgentTriage(ticketsDir, id, verbose)`
  from `cmdAgentTriage` (which became a thin wrapper). Added `maybeAutoTriage`
  which silently skips if no config or `AutoTriage=false`, and prints a
  non-fatal warning to stderr on failure.

- **`cmd_create.go`** — Calls `maybeAutoTriage` after `EmitMutationEvent` when
  `--triage` was set.

- **`cmd_update.go`** — Same pattern after its `EmitMutationEvent`. Covers
  `ko update --triage` and the `ko triage <id> <instructions>` alias (which
  routes through `cmdUpdate`).

- **`specs/ticket_triage.feature`** — 4 new scenarios: auto-trigger on create,
  auto-trigger on update, no-trigger when `auto_triage` is absent, and
  non-fatal failure.

- **`testdata/ticket_triage/`** — 4 new txtar integration tests, all passing.

## Notable Decisions

- **Non-fatal failure**: Per the decision note, auto-triage failure does not
  prevent the ticket from being saved. The triage field remains set, a warning
  is printed to stderr, and the user can retry manually with
  `ko agent triage <id>`.

- **Condition is `--triage != ""`**: Auto-triage fires only when `--triage` was
  explicitly passed, not on every create/update. This matches the ticket's
  stated trigger: "when a ticket is created or modified to have triage
  instructions."

## Known Issue / Follow-Up Needed

`pipeline.go` was already at 609 lines (over the 500-line invariant) before
this change and is now at 613 lines. INVARIANTS.md states "don't let new work
make them bigger." The addition was minimal (4 lines) and was acknowledged in
the implementation plan as a known invariant conflict. A separate ticket should
be created to split `pipeline.go` along behavioral seams.
