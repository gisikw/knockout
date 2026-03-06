Plan written to `.ko/tickets/ko-c482.artifacts/plan.md`. Here's the summary:

**Goal:** Add `actionable bool` to `agentStatusJSON` — `true` when `ReadyQueue` or `TriageQueue` is non-empty.

**Key findings:**
- `agentStatusJSON` struct and `cmdAgentStatus` are in `cmd_agent.go:265`
- `ReadyQueue` and `TriageQueue` already exist in `loop.go` and do exactly what the ticket needs
- The field should be computed after provisioning is confirmed, errors treated as empty (safe default)

**3 tasks:**
1. Add `Actionable bool \`json:"actionable"\`` to `agentStatusJSON`
2. Populate it in `cmdAgentStatus` using `ReadyQueue` + `TriageQueue` after `status.Provisioned = true`
3. Extend `TestCmdAgentStatusJSON` with cases for ready tickets, triageable tickets, and neither

No open questions — the logic is fully defined by existing functions.
