# Implementation Summary: ko agent report

## What Was Done

Successfully implemented the `ko agent report` command that displays summary statistics from the last agent loop run. The command reads the `.ko/agent.log` file, extracts the last JSONL summary line, and presents it in either human-readable or JSON format.

## Implementation Details

### Files Modified

1. **cmd_agent.go**
   - Added `case "report"` routing in `cmdAgent()` switch statement
   - Implemented `cmdAgentReport()` function that:
     - Resolves project tickets directory
     - Reads `.ko/agent.log` and scans for JSONL lines (starting with `{`)
     - Extracts the last JSONL line
     - Parses it into `agentReportJSON` struct
     - Outputs human-readable format or JSON (with `--json` flag)
     - Handles missing log file gracefully with "no runs found"
   - Added `agentReportJSON` struct matching the JSONL summary format from `writeAgentLogSummary()`

2. **cmd_agent_test.go**
   - Added `TestCmdAgentReportJSON()` with comprehensive test cases:
     - No log file (returns empty result)
     - No JSONL lines in log (returns empty result)
     - Single JSONL line (parsed correctly)
     - Multiple JSONL lines (last one is selected)
   - All test cases verify both structure and field values

3. **testdata/loop/loop_report.txtar**
   - Created integration test verifying end-to-end behavior:
     - Runs `ko agent loop --max-tickets 2`
     - Verifies `ko agent report` displays all expected fields
     - Verifies `ko agent report --json` outputs valid JSON
     - Uses existing loop test infrastructure with fake-llm

4. **README.md**
   - Added `agent report` to the command list (line 46)
   - Added to JSON output section (line 73) with description

### Verification

✅ All unit tests pass (`TestCmdAgentReportJSON`)
✅ All integration tests pass (`TestLoop/loop_report`)
✅ Command works in both human-readable and JSON modes
✅ Follows existing patterns for agent subcommands
✅ No deviations from the plan

## Notable Decisions

1. **JSONL Line Extraction**: The implementation scans all lines and keeps the last one starting with `{`. This is simple and robust, handling the mixed-format log file correctly.

2. **Empty Result Handling**: When no log file exists or no JSONL lines are found, the command returns an empty JSON object `{}` in JSON mode, or "no runs found" in human-readable mode. This matches the pattern used by other agent commands.

3. **Struct Definition**: The `agentReportJSON` struct mirrors the exact field names and types used in `writeAgentLogSummary()`, ensuring consistent serialization/deserialization.

## Compliance with INVARIANTS.md

### ⚠️ Missing Spec File

**Issue**: INVARIANTS.md states "Every behavior has a spec" and "Every spec has a test." While the implementation has comprehensive tests (`TestCmdAgentReportJSON` and `testdata/loop/loop_report.txtar`), there is **no corresponding spec file** in `specs/`.

**Context**:
- No existing spec file covers agent commands (`start`, `stop`, `status`, `report`)
- `specs/loop.feature` covers the loop behavior but not the reporting
- `specs/agent_harnesses.feature` covers harness architecture, not CLI commands

**Recommendation**: A new spec file should be created (e.g., `specs/agent_commands.feature`) to document the behavioral contracts for all agent CLI commands, including `status`, `start`, `stop`, and `report`.

### Other Invariants

✅ **Build**: No runtime dependencies added
✅ **Data Model**: Uses existing JSONL format from agent.log
✅ **Code Organization**: Decision logic is pure (JSONL parsing), I/O is in cmdAgentReport()
✅ **File Size**: All modified files remain under 500 lines
✅ **Error Handling**: CLI errors go to stderr with non-zero exit code
✅ **Naming**: Function follows existing `cmdAgent<Subcommand>()` pattern
✅ **Tests**: Unit tests in `cmd_agent_test.go`, integration test in `testdata/loop/`

## Future Work

1. Create `specs/agent_commands.feature` to document agent CLI command behaviors
2. Consider adding `--format` flag for additional output formats (table, CSV) if needed
3. Consider adding `--all` flag to show history of all runs (not just the last one)
