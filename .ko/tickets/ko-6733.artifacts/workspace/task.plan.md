All done. Here's a summary of what was implemented:

**`cmd_agent_triage.go`** (new file) — `cmdAgentTriage`:
- Resolves ticket via `resolveProjectTicketsDir` + `ResolveTicket`
- Fails with clear error if `t.Triage == ""`
- Requires pipeline config via `FindPipelineConfig` + `LoadPipeline`; fails if absent
- Constructs prompt: `## Ticket` + ticket content + `## Instructions` + triage string
- Forces `allowAll=true` regardless of pipeline config (per ticket decision)
- Injects `TICKETS_DIR`, `KO_TICKET_WORKSPACE`, `KO_ARTIFACT_DIR` env vars
- Uses `parseTimeout(p.StepTimeout)` with context timeout
- On success: reloads ticket, clears `Triage`, saves, prints `"<id>: triage cleared"`

**`cmd_agent.go`** — Added `case "triage": return cmdAgentTriage(args[1:])` and updated usage string.

**`main.go`** — Added `agent triage <id>` to `ko help` output.

**`specs/ticket_triage.feature`** — Added 3 new scenarios: success case, no-triage error, no-config error.

**`testdata/agent_triage/`** — 3 txtar tests covering: success + triage cleared, no-triage failure, no-config failure.
