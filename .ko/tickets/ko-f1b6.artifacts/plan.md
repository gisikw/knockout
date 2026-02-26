## Goal
Add `ko agent report` command that displays summary statistics from the last agent loop run.

## Context
The agent loop already writes JSONL summary lines to `.ko/agent.log` after each run (implemented in `cmd_loop.go:writeAgentLogSummary()`). Each summary line contains:
- `ts`: timestamp in RFC3339 format
- `tickets_processed`: total tickets attempted
- `succeeded`, `failed`, `blocked`, `decomposed`: outcome counts
- `stop_reason`: why the loop stopped ("empty", "max_tickets", "max_duration", "signal", "build_error")
- `runtime_seconds`: elapsed time for the run

The `.ko/agent.log` file contains mixed format: human-readable text output ("loop: building ko-xxx") interspersed with JSONL summary lines at the end of each run.

Existing agent subcommands are implemented in `cmd_agent.go` and follow the pattern:
- Subcommand routing in `cmdAgent()`
- Individual `cmdAgent<Subcommand>()` functions
- Support for `--json` flag for machine-readable output
- Helper functions like `agentPidPath()`, `agentLogPath()` for path construction

Tests follow two patterns:
- Unit tests in `cmd_agent_test.go` (like `TestCmdAgentStatusJSON`)
- Integration tests in `testdata/loop/*.txtar` using testscript

## Approach
Add a new `ko agent report` subcommand that reads `.ko/agent.log`, extracts the last JSONL summary line, and displays it in human-readable format (or JSON with `--json` flag). The command will parse the most recent summary line and present statistics about the last agent loop run.

## Tasks
1. [cmd_agent.go:cmdAgent] — Add `case "report"` to the subcommand switch, routing to `cmdAgentReport()`.
   Verify: `ko agent` help output mentions report subcommand.

2. [cmd_agent.go] — Add `cmdAgentReport()` function that:
   - Resolves project tickets directory
   - Reads `.ko/agent.log` file
   - Extracts the last JSONL line (lines starting with `{`)
   - Parses it into a struct matching the summary format
   - Outputs human-readable summary (or JSON with `--json`)
   - Handles missing log file gracefully ("no runs found")
   Verify: `ko agent report` displays stats; `ko agent report --json` outputs valid JSON.

3. [cmd_agent_test.go] — Add `TestCmdAgentReportJSON()` unit test verifying:
   - Report with no log file returns appropriate message
   - Report with valid JSONL line parses and displays correctly
   - `--json` flag produces valid JSON output
   - Multiple JSONL lines (last one is selected)
   Verify: `go test -run TestCmdAgentReportJSON` passes.

4. [testdata/loop/loop_report.txtar] — Add integration test verifying:
   - After running `ko agent loop`, `ko agent report` displays summary
   - Output contains all expected fields (processed, succeeded, failed, etc.)
   - `--json` flag works in integration context
   Verify: `go test -run TestLoop/loop_report` passes.

5. [README.md] — Update agent commands section to document `ko agent report`:
   - Add to command list around line 46
   - Add to JSON output section around line 73
   - Brief description: "Show summary statistics from the last agent loop run"
   Verify: Documentation is consistent with other agent commands.

## Open Questions
None — the implementation is straightforward. The JSONL format is already established and stable, and the command follows existing patterns for agent subcommands.
