Perfect! All tasks completed successfully. Here's a summary of what was implemented:

## Summary

I successfully implemented pipeline support for the `resolved` status, enabling workflows to configure whether successful builds land tickets in `resolved` or `closed` status. Here's what was changed, organized by task:

### Task 1: Added OnSuccess field to Workflow struct
- **File**: `workflow.go:46`
- Added `OnSuccess string` field to the `Workflow` struct to store the desired status on successful completion ("closed" or "resolved")

### Task 2: Parse on_success field in ParsePipeline
- **File**: `pipeline.go:199`
- Added parsing for `on_success` at the workflow level in the YAML parser
- **File**: `pipeline_test.go:733`
- Added test case `TestParsePipelineWorkflowOnSuccess` to verify parsing of both `resolved` and `closed` values

### Task 3: Update runWorkflow return signature
- **File**: `build.go:151`
- Changed `runWorkflow` signature from `(Outcome, error)` to `(Outcome, string, error)` to return the final workflow name
- Updated all return statements in `runWorkflow` to include the workflow name
- Updated `applyDisposition` to also return the workflow name and propagate it from recursive calls

### Task 4: Update RunBuild to use workflow OnSuccess config
- **File**: `build.go:117-133`
- Modified `RunBuild` to capture the final workflow name from `runWorkflow`
- Added logic to look up the final workflow's `OnSuccess` config
- Sets status to `resolved` if configured, otherwise defaults to `closed` for backward compatibility

### Task 5: Add resolved to validDispositions
- **File**: `disposition.go:19-26`
- Added `"resolved"` to the `validDispositions` map
- **File**: `disposition.go:127-132`
- Updated `DispositionSchema` to document the new `resolved` disposition with an example

### Task 6: Add resolved case to applyDisposition
- **File**: `build.go:313-319`
- Added case for `"resolved"` disposition that:
  - Adds a note to the ticket
  - Sets status to `resolved`
  - Returns `OutcomeFail` to halt the workflow (resolved tickets shouldn't continue building)

### Task 8: Add on_success to research workflow in example
- **File**: `examples/structured/pipeline.yml:42`
- Added `on_success: resolved` to the `research` workflow as a demonstration

All changes follow the plan exactly, maintain backward compatibility (empty/missing `on_success` defaults to `closed`), and all tests pass successfully. The implementation provides both workflow-level configuration and a disposition override, allowing flexible control over how successful builds conclude.
