Perfect! The implementation is complete. Let me provide a summary of the changes made:

## Summary

I have successfully implemented granular tool permissions in the pipeline configuration as specified in the plan. Here's a summary of the changes made, organized by task:

### Core Data Model Changes
1. **Pipeline struct** (pipeline.go): Added `AllowedTools []string` field
2. **Workflow struct** (workflow.go): Added `AllowedTools []string` field  
3. **Node struct** (workflow.go): Added `AllowedTools []string` field

### YAML Parser Updates
4. **pipeline.go**: Updated the YAML parser to handle `allowed_tools` at all three levels (pipeline, workflow, node) with support for both inline `[a, b]` and multiline list formats

### Resolution Logic
5. **build.go**: Created `resolveAllowedTools()` function with override semantics (node > workflow > pipeline)
6. **build.go**: Updated `runWorkflow()` and `runNode()` to resolve and pass allowed tools to the adapter

### Adapter Interface Updates
7. **adapter.go**: Updated `AgentAdapter` interface signature to accept `allowedTools []string` parameter
8. **adapter.go**: Updated `RawCommandAdapter.BuildCommand()` signature (no implementation for raw commands)
9. **harness.go**: Updated `TemplateAdapter.BuildCommand()` to format tools as comma-separated values and expand `${allowed_tools}` template variable to `--allowed-prompts\n<tools-csv>`

### Harness Configuration
10. **agent-harnesses/claude.yaml**: Added `${allowed_tools}` template variable
11. **agent-harnesses/cursor.yaml**: Added `${allowed_tools}` for forward compatibility

### Tests
12. **pipeline_test.go**: Added `TestParsePipelineAllowedTools()` validating parsing at all levels in both inline and multiline formats
13. **pipeline_test.go**: Added `TestResolveAllowedToolsOverride()` validating override semantics across the three levels
14. **harness_test.go**: Updated all existing test calls to include the new `allowedTools` parameter

### Documentation
15. **README.md**: Added `allowed_tools` row to pipeline options table with complete description
16. **README.md**: Added `allowed_tools` entry to node properties table

All tests pass successfully, and the implementation follows the override semantics decision (node-level completely replaces parent lists). The feature integrates cleanly with the existing `allow_all_tool_calls` boolean, with `allowed_tools` only being used when `allow_all_tool_calls` is false.
