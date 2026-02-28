The plan is written. Here's a summary:

## Plan Summary

**Goal**: Add `ko agent triage <id>` that invokes the configured model with the ticket's triage instruction, then clears the triage field on success.

**Key findings**:
- `cmd_agent.go` is at 457 lines — adding here would cross the 500-line invariant, so the implementation goes in a new `cmd_agent_triage.go` file (matching the `cmd_build.go` / `cmd_loop.go` pattern)
- "Ko usage tokens" = `TICKETS_DIR`, `KO_TICKET_WORKSPACE`, `KO_ARTIFACT_DIR` env vars that let the spawned model run `ko` commands and find the right project
- Prompt structure mirrors `runPromptNode`: ticket content + `## Instructions` with the triage string
- On success: reload ticket (model may have modified it), clear `Triage`, save

**Tasks**:
1. New `cmd_agent_triage.go` — full `cmdAgentTriage` implementation
2. `cmd_agent.go` — add `triage` case + usage string
3. `main.go` — add to `ko help` text
4. `specs/ticket_triage.feature` — new spec scenarios
5. `testdata/agent_triage/triage_run.txtar` — integration test with mock harness
6. `testdata/agent_triage/triage_no_triage.txtar` — error case test

**Open Questions**:
1. Should `allowAll` be forced to `true` for triage (ergonomic but bypasses config), or read from pipeline config?
2. Should pipeline config be required, or fall back to a default claude adapter?
