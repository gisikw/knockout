# Implementation Summary: Granular Tool Permissions

## What Was Done

Successfully implemented granular tool permissions in the pipeline configuration system, allowing users to specify which tools are auto-allowed at pipeline, workflow, or node levels instead of the all-or-nothing `allow_all_tool_calls` flag.

### Changes Made

**Data Model** (3 files modified):
- Added `AllowedTools []string` field to `Pipeline`, `Workflow`, and `Node` structs
- Enables three-level configuration hierarchy with override semantics

**YAML Parsing** (pipeline.go):
- Updated parser to handle `allowed_tools` at all three levels
- Supports both inline `[Tool1, Tool2]` and multiline list syntax
- Follows existing patterns for parsing list fields (routes, skills)

**Resolution Logic** (build.go):
- Created `resolveAllowedTools()` function with override semantics (node > workflow > pipeline)
- Updated `runWorkflow()` and `runNode()` to resolve and thread allowed tools through execution
- Returns nil when no tools specified at any level

**Adapter System** (adapter.go, harness.go):
- Updated `AgentAdapter` interface signature to accept `allowedTools []string` parameter
- Modified `TemplateAdapter.BuildCommand()` to format tools as comma-separated values
- Template variable `${allowed_tools}` expands to `--allowed-prompts\nRead,Write,Bash` when set, empty string when nil/empty
- Updated `RawCommandAdapter` signature (no implementation, as raw commands don't support this)

**Agent Harnesses**:
- Added `${allowed_tools}` to claude.yaml (expands to `--allowed-prompts` flag)
- Added `${allowed_tools}` to cursor.yaml (forward compatibility)

**Tests** (pipeline_test.go, harness_test.go):
- `TestParsePipelineAllowedTools`: Validates parsing at all levels in both formats
- `TestResolveAllowedToolsOverride`: Validates override semantics including empty list case
- Updated all existing harness tests to pass new parameter

**Documentation** (README.md):
- Added pipeline option row with complete description of semantics
- Added node property documentation
- Documented case-sensitivity and interaction with `allow_all_tool_calls`

## Key Decisions Made During Implementation

1. **Override Semantics Confirmed**: As decided in ticket notes (2026-02-24 13:02:23 UTC), used override semantics where node-level completely replaces parent lists, matching `allow_all_tool_calls` behavior.

2. **Empty List Handling**: An empty list `[]` at a given level means "allow no tools" at that level. To inherit from parent, omit the field (nil value). This provides explicit control over restriction.

3. **Tool Format**: Tools are formatted as comma-separated values for the `--allowed-prompts` flag, matching Claude CLI expectations.

4. **Template Variable Expansion**: When tools list is non-empty, expands to `--allowed-prompts\n<tools>`. The newline ensures the flag and value are on separate lines in the template args list.

## Test Results

All tests pass:
- New tests: `TestParsePipelineAllowedTools`, `TestResolveAllowedToolsOverride`
- All existing tests continue passing with updated signatures
- Full test suite: `ok git.gisi.network/infra/knockout 3.191s`

## Notable Implementation Details

1. **Nil vs Empty**: The implementation correctly distinguishes between `nil` (not set, inherit from parent) and `[]` (empty list, allow no tools). This is critical for override semantics.

2. **Integration with Existing System**: The feature integrates cleanly with existing `allow_all_tool_calls`. When `allow_all_tool_calls: true`, it takes precedence and skips all permission checks. `allowed_tools` is only consulted when `allow_all_tool_calls` is false or unset.

3. **Parsing Pattern Consistency**: Used the same parsing pattern as existing list fields (routes, skills), maintaining codebase consistency.

4. **Forward Compatibility**: Added the template variable to cursor.yaml even though Cursor may not support the flag yet, ensuring the configuration is ready when support is added.

## Compliance

✅ All 14 planned tasks completed
✅ All open questions resolved before implementation
✅ Tests added and passing
✅ Documentation updated
✅ No deviations from plan
✅ No invariant violations
✅ Override semantics correctly implemented per decision
