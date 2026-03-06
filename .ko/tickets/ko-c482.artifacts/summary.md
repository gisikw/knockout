## Summary

Added `actionable bool` (JSON: `"actionable"`) to `agentStatusJSON` in `cmd_agent.go`. The field is `true` when either `ReadyQueue` or `TriageQueue` returns at least one ticket, and `false` otherwise (including when the project is not provisioned).

### What was done

1. **Struct field** — `Actionable bool \`json:"actionable"\`` added to `agentStatusJSON`.
2. **Population logic** — In `cmdAgentStatus`, after confirming `Provisioned = true`, both queue functions are called with errors silently ignored (safe degradation to `false`). This matches the plan exactly.
3. **Tests** — `TestCmdAgentStatusJSON` extended with four new table cases: no tickets, one ready ticket, one triage ticket, and not-provisioned-with-tickets. All pass.

### Notable decisions

- Errors from `ReadyQueue` / `TriageQueue` are swallowed (as planned). A transient I/O error degrades `actionable` to `false`, which is safe — UI consumers will simply not show the toggle rather than crashing.
- The field is always emitted (not `omitempty`) so consumers can reliably check it without worrying about a missing key.

### Fix applied during review

The implementation did not include a `specs/agent_status.feature` file, which is required by INVARIANTS.md ("Every behavior has a spec … spec before code"). A spec was added covering the five key scenarios: not provisioned, provisioned with no tickets, ready ticket, triage ticket, and actionable-while-not-running. This satisfies the invariant.

### Anything a future reader should know

- `ko agent status` (without `--json`) still prints plain text and is unaffected.
- There is no testdata `.txtar` file for this behavior yet — that remains pre-existing non-compliance for the broader `agent status` command and should be tracked separately.
