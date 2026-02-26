Perfect! I've created a plan for implementing the `ko agent report` command. Here's a summary of what the plan covers:

## Summary

**Goal**: Add a `ko agent report` command that displays summary statistics from the last agent loop run.

**Key Insights**:
- The infrastructure is already in place — `writeAgentLogSummary()` writes JSONL summary lines to `.ko/agent.log` after each loop run
- The new command just needs to read the last JSONL line and present it in a user-friendly format
- Follows existing patterns for agent subcommands (routing, `--json` support, path helpers)

**Implementation Plan**:
1. Add routing for the new `report` subcommand in `cmd_agent.go`
2. Implement `cmdAgentReport()` to read and parse the last JSONL summary line
3. Add unit tests in `cmd_agent_test.go`
4. Add integration test in `testdata/loop/loop_report.txtar`
5. Update README.md documentation

**No open questions** — this is a straightforward implementation that leverages existing infrastructure and follows established patterns.
