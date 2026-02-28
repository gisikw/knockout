All tests pass. Here's a summary of what was already implemented (from the previous build attempt) and verified:

## Summary

All 6 tasks from the plan were already implemented in the previous build attempt. I verified each:

**Task 1 — `cmd_agent_triage.go` (new file)**
- `cmdAgentTriage` resolves the ticket, validates `t.Triage != ""`, loads pipeline config (fails if absent), constructs the prompt, forces `allowAll=true`, injects `TICKETS_DIR`/`KO_TICKET_WORKSPACE`/`KO_ARTIFACT_DIR`, runs with timeout, and on success reloads ticket, clears triage, saves, prints `"<id>: triage cleared"`.

**Task 2 — `cmd_agent.go`**
- `case "triage": return cmdAgentTriage(args[1:])` added to the switch.
- `"  triage <id>  Run triage instructions against a ticket"` added to the usage string.

**Task 3 — `main.go:cmdHelp`**
- `"  agent triage <id>  Run triage instructions against a ticket"` added to the help output.

**Task 4 — `specs/ticket_triage.feature`**
- 3 new scenarios added: success case, no-triage failure, no-config failure.

**Task 5 — `testdata/agent_triage/triage_run.txtar`**
- Mock harness exits 0; asserts exit 0, output contains `"triage cleared"`, ticket no longer contains `triage:`.

**Task 6 — `testdata/agent_triage/triage_no_triage.txtar`** + **`triage_no_config.txtar`**
- No-triage and no-config failure cases tested.

`go build ./...` passes. All tests pass (`TestAgentTriage`: 3/3, full suite: ok).
