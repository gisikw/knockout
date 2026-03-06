All 8 tests pass. Summary:

**Task 1** — Added `Actionable bool \`json:"actionable"\`` to `agentStatusJSON` in `cmd_agent.go:270`.

**Task 2** — After `status.Provisioned = true` in `cmdAgentStatus`, call `ReadyQueue` and `TriageQueue` (errors silently ignored) and set `status.Actionable = len(ready) > 0 || len(triage) > 0`. The value flows into all existing JSON output paths.

**Task 3** — Extended `TestCmdAgentStatusJSON` with 4 new cases: provisioned/no tickets (`false`), provisioned/ready ticket (`true`), provisioned/triage ticket (`true`), not provisioned/with tickets (`false`). Added a `tickets []string` field to the test struct for writing fixture `.md` files. All 8 cases pass.

No deviations from the plan.
