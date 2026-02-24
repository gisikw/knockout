# Implementation Summary: Pipeline Support for Resolved Status

## What Was Done

Successfully implemented dual-layer support for the `resolved` status in the pipeline system:

1. **Workflow-level configuration**: Added `on_success: resolved|closed` field to workflow definitions
   - Parsed in `pipeline.go` (line 197-198)
   - Stored in `Workflow.OnSuccess` field (workflow.go:46)
   - Defaults to empty string (backward compatible → "closed")

2. **Final workflow tracking**: Modified `runWorkflow()` to return the name of the terminal workflow
   - Changed signature from `(Outcome, error)` to `(Outcome, string, error)`
   - Tracks which workflow completed last (handles routing correctly)
   - Updated all call sites in build.go

3. **Success outcome routing**: Updated `RunBuild()` to respect final workflow's `on_success` config
   - Looks up the final workflow's `OnSuccess` field after success
   - Sets ticket to `resolved` if configured, otherwise `closed`
   - Preserves backward compatibility (empty → "closed")

4. **Disposition override**: Added new `resolved` disposition type
   - Added to `validDispositions` map (disposition.go:25)
   - Documented in DispositionSchema with example (disposition.go:129-132)
   - Implemented in `applyDisposition()` (build.go:313-320)
   - Returns `OutcomeFail` to halt workflow (resolved tickets don't continue building)

5. **Test coverage**: Added comprehensive test
   - `TestParsePipelineWorkflowOnSuccess` validates parsing of all three cases
   - All existing tests continue to pass

6. **Example**: Updated `examples/structured/pipeline.yml` to demonstrate usage
   - Added `on_success: resolved` to the `research` workflow

## Key Design Decisions

- **Backward compatibility**: Empty/missing `on_success` defaults to `closed` behavior
- **Terminal workflow wins**: Only the final workflow's config matters (not intermediate routed workflows)
- **Disposition is terminal**: Like `fail` and `blocked`, the `resolved` disposition halts the workflow immediately
- **Note formatting**: Resolved disposition adds note: `"ko: RESOLVED at node '<name>'"` with optional reason

## Notable Implementation Details

- The `resolved` disposition returns `OutcomeFail` (not a new outcome type) because it should halt execution
- The workflow name tracking threads through recursive `runWorkflow()` calls properly
- The final workflow name is returned as empty string `""` for non-route dispositions (continue, fail, blocked, decompose)
- When routing, the recursive call returns the routed workflow's final name

## Testing

All tests pass:
- New test validates parsing of `on_success` at workflow level
- Existing build tests continue to pass with new signature
- `go build` succeeds

## Future Considerations

This implementation enables research workflows to default to `resolved` status, requiring human review before closing. Task workflows can continue defaulting to `closed`. Decision nodes can override either default using the `resolved` disposition when appropriate.
