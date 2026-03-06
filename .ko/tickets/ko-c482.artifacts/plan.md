## Goal
Add an `actionable` boolean field to `ko agent status --json` output that is `true` when the ready queue is non-empty or triageable tickets exist.

## Context
- `agentStatusJSON` struct and `cmdAgentStatus` live in `cmd_agent.go` (lines 265–349).
- `ReadyQueue(ticketsDir string) ([]string, error)` in `loop.go:94` returns IDs of tickets ready to build (open/in_progress, deps resolved, not snoozed, no triage field).
- `TriageQueue(ticketsDir string) ([]*Ticket, error)` in `loop.go:44` returns tickets with a non-empty `triage` field.
- Both functions are already tested and available for reuse.
- Existing tests in `cmd_agent_test.go` use a table-driven pattern, capturing stdout via `os.Pipe()` and unmarshaling `agentStatusJSON`.
- The `actionable` field should be populated regardless of whether the agent is running — consumers use it to decide whether to show the toggle at all.
- When the project is not provisioned (no pipeline config), `actionable` stays `false` (consistent with the rest of the struct being zero-valued in that branch).

## Approach
Add `Actionable bool \`json:"actionable"\`` to `agentStatusJSON`. In `cmdAgentStatus`, after confirming the project is provisioned, call `ReadyQueue` and `TriageQueue` and set `status.Actionable = len(readyQueue) > 0 || len(triageQueue) > 0`. The field is computed once and flows into all existing JSON output paths without restructuring them.

## Tasks
1. [`cmd_agent.go:agentStatusJSON`] — Add `Actionable bool \`json:"actionable"\`` field to the struct.
   Verify: `go build ./...` succeeds.

2. [`cmd_agent.go:cmdAgentStatus`] — After setting `status.Provisioned = true`, call `ReadyQueue` and `TriageQueue`. Set `status.Actionable = len(ready) > 0 || len(triage) > 0`. Errors from both calls should be silently ignored (treat as empty queue) — `actionable` degrading to `false` on I/O error is safe for UI consumers.
   Verify: `go build ./...` succeeds; manual `ko agent status --json` in a project with an open ticket shows `"actionable":true`.

3. [`cmd_agent_test.go`] — Extend `TestCmdAgentStatusJSON` with cases that cover: (a) provisioned, no tickets → `actionable: false`; (b) provisioned, one ready ticket → `actionable: true`; (c) provisioned, one ticket with triage field → `actionable: true`; (d) not provisioned → `actionable: false`. Each case writes minimal ticket `.md` files into the temp `ticketsDir`.
   Verify: `go test ./... -run TestCmdAgentStatusJSON` passes.

## Open Questions
None. The definition of "ready" and "triageable" is already encoded in `ReadyQueue` and `TriageQueue`; reusing them is the correct approach.
