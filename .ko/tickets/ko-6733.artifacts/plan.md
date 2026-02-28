## Goal
Add `ko agent triage <id>` subcommand that invokes the configured model with the ticket's triage instructions and clears the triage field on success.

## Context

**Triage field** (`Ticket.Triage`, `ticket.go:35`): a free-text string set via `ko triage <id> <instructions>` or `ko update --triage`. Tickets with a non-empty triage value are excluded from `ko ready` and listed by `ko triage` (no args).

**Agent subcommand dispatch** is in `cmd_agent.go:cmdAgent`. Each subcommand is a standalone function; `cmdAgentBuild` lives in `cmd_build.go`, `cmdAgentLoop`/`cmdAgentStart` in `cmd_loop.go`, etc. The pattern is: one file per major subcommand. `cmd_agent.go` is 457 lines — adding 70+ lines would exceed the 500-line invariant, so the new function goes in a new file `cmd_agent_triage.go`.

**Adapter/harness** (`harness.go`, `adapter.go`): `p.Adapter()` returns an `AgentAdapter` that builds an `exec.Cmd`. The shell harness passes `KO_PROMPT`, `KO_MODEL`, `KO_SYSTEM_PROMPT`, `KO_ALLOW_ALL`, `KO_ALLOWED_TOOLS` as env vars. `runPromptNode` in `build.go:441` shows the full prompt-construction + env-setup pattern (including appending `KO_TICKET_WORKSPACE`, `KO_ARTIFACT_DIR`, `KO_BUILD_HISTORY`).

**"Ko usage tokens"**: The ticket means providing the environment variables that let the spawned model operate on tickets:
- `TICKETS_DIR=<ticketsDir>` — lets `ko` commands in the model's shell find the right project without needing cwd
- `KO_ARTIFACT_DIR=<artifactDir>` — artifact directory for the model to write to
- `KO_TICKET_WORKSPACE=<wsDir>` — workspace dir for the model

**Prompt structure** matches `runPromptNode` (build.go:463–487): `## Ticket`, ticket content, `## Instructions`, then the triage string as instructions. No discretion guidance, no disposition schema (this is an action, not a decision node).

**Clearing triage**: after the command succeeds, reload the ticket (model may have modified it), set `Triage = ""`, call `SaveTicket`. Reload-then-clear prevents clobbering model changes.

**Pipeline requirement**: Loading the pipeline is needed to get the adapter, model, allowAll, and allowedTools. If no pipeline config exists, the command fails with a clear error (the command is meaningless without knowing which agent to use).

**Timeout**: use `p.StepTimeout` via `parseTimeout`; defaults to 15 minutes if unset.

## Approach

Add `cmdAgentTriage` in a new `cmd_agent_triage.go` file. It: resolves the ticket, validates `Triage != ""`, loads pipeline config, constructs a prompt (ticket content + triage as instructions), invokes the adapter with `TICKETS_DIR`/workspace env vars set, and on success clears the triage field. Wire it into `cmdAgent` and update help strings. Add a spec to `specs/ticket_triage.feature` and a txtar integration test with a mock harness.

## Tasks

1. **[cmd_agent_triage.go — new file]** Implement `cmdAgentTriage(args []string) int`:
   - Parse args via `flag.FlagSet` with `--verbose`/`-v` flag; require exactly one positional arg (ticket ID).
   - Resolve ticketsDir via `resolveProjectTicketsDir`, then resolve ticket ID via `ResolveTicket`.
   - Load ticket; error if `t.Triage == ""` ("ticket has no triage value").
   - Load pipeline via `FindPipelineConfig` + `LoadPipeline`; error if not found.
   - Construct prompt: `"## Ticket\n\n# <title>\n<body>\n\n## Instructions\n\n<triage>"`.
   - Ensure artifact dir via `EnsureArtifactDir`; create workspace via `CreateWorkspace`.
   - Call `p.Adapter().BuildCommand(prompt, p.Model, "", p.AllowAll, p.AllowedTools)`.
   - Wrap the resulting cmd in a context with timeout from `parseTimeout(p.StepTimeout)`.
   - Append to cmd env: `TICKETS_DIR=<ticketsDir>`, `KO_TICKET_WORKSPACE=<wsDir>`, `KO_ARTIFACT_DIR=<artifactDir>`.
   - Run cmd (verbose: stream to stdout/stderr; non-verbose: capture output, stream stderr).
   - On error: print stderr, return 1.
   - On success: reload ticket, set `t.Triage = ""`, `SaveTicket`, print `"<id>: triage cleared"`, return 0.
   - Verify: `go build ./...` passes; manually test error paths (no triage, no config).

2. **[cmd_agent.go:cmdAgent]** Add `case "triage": return cmdAgentTriage(args[1:])` in the switch, and add `"  triage <id>    Run triage instructions against a ticket"` to the usage string in the error path.
   Verify: `ko agent` with no args shows `triage` in usage.

3. **[main.go:cmdHelp]** Add `"  agent triage <id>   Run triage instructions against a ticket"` to the help text alongside the other `agent` subcommands.
   Verify: `ko help` shows the new subcommand.

4. **[specs/ticket_triage.feature]** Add new scenarios:
   - `ko agent triage <id>` succeeds and clears triage field (using mock harness).
   - `ko agent triage <id>` with no triage set fails with error.
   Verify: spec is self-consistent with the implementation.

5. **[testdata/agent_triage/triage_run.txtar — new dir/file]** Write a txtar test with a mock shell harness that exits 0. Assert: exit 0, output contains "triage cleared", ticket file no longer contains `triage:` in frontmatter.
   Verify: `go test ./... -run TestScript/agent_triage` passes.

6. **[testdata/agent_triage/triage_no_triage.txtar]** Test: ticket has no triage field → command fails with "has no triage value".
   Verify: test passes.

## Open Questions

1. **`allowAll` behavior**: The plan uses `p.AllowAll` from pipeline config. In practice, triage operations ("unblock this ticket", "break this apart") require running `ko` shell commands, which needs `allowAll=true` or `Bash` in allowed tools. Should `ko agent triage` force `allowAll=true` regardless of pipeline config, or should users be expected to set `allow_all_tool_calls: true` in their pipeline? Forcing `true` is more ergonomic but bypasses user permission settings. **Decision needed.**

2. **Pipeline required or optional**: The plan requires a pipeline config. If no pipeline config exists, should the command fall back to a hardcoded default adapter ("claude") and empty model? This would make the command work even on freshly initialized projects. Given that the ticket says "load default model," a pipeline is probably expected, but it's worth confirming.
