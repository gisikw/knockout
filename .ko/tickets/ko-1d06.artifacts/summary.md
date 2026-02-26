# Implementation Summary: on_loop_complete hooks

## What Was Done

Successfully implemented post-loop hooks that run once after the agent loop completes, allowing for general tasks like updating the README or pushing changes to be done just once after all tickets are processed.

## Implementation Details

All 8 planned tasks were completed exactly as specified:

### 1. Pipeline Struct (pipeline.go:35-38)
Added `OnLoopComplete []string` field to the Pipeline struct after the existing OnClose field, with clear documentation that these hooks run after the loop completes.

### 2. YAML Parsing (pipeline.go:142, 190-193, 361-366)
Added complete parsing support for the `on_loop_complete:` section in pipeline YAML files, following the same pattern as the existing hook types.

### 3. Hook Execution Function (loop.go:142-188)
Created `runLoopHooks()` helper function that:
- Accepts loop result and elapsed duration
- Sets comprehensive environment variables: `LOOP_PROCESSED`, `LOOP_SUCCEEDED`, `LOOP_FAILED`, `LOOP_BLOCKED`, `LOOP_DECOMPOSED`, `LOOP_STOPPED`, `LOOP_RUNTIME_SECONDS`
- Expands variables in hook commands
- Executes hooks in the project root with proper environment

### 4. Hook Invocation (cmd_loop.go:218-222)
Integrated hook execution into the loop command immediately after writing the agent log summary but before cleanup. Hook failures are logged to stderr but don't affect the loop exit code, as designed.

### 5. Specification (specs/loop.feature:88-105)
Added two comprehensive scenarios:
- "on_loop_complete hooks run after loop finishes" — validates all environment variables are set correctly
- "on_loop_complete hooks run regardless of stop reason" — validates hooks run for different stop reasons (empty, max_tickets, etc.)

### 6. Test (testdata/loop/loop_on_complete_hook.txtar)
Created a complete txtar test that:
- Sets up a pipeline with on_loop_complete hooks
- Writes all env vars to a file for verification
- Validates hook execution and environment variable correctness
- **Test passes:** ✅

### 7. Example (examples/default/pipeline.yml:40-46)
Added a commented-out example showing:
- All available environment variables
- Conditional logic pattern (only push if succeeded > 0)
- Clear documentation for users

### 8. Documentation (README.md:315-319)
Updated the Hooks section with complete documentation:
- When hooks run (after loop completes, regardless of stop reason)
- Available environment variables
- Error handling behavior (logged but don't affect exit code)

## Notable Decisions

1. **Environment Variables**: Exposed all loop result fields as env vars to enable intelligent hook behavior (e.g., only push if at least one ticket succeeded)

2. **Error Handling**: Hook failures are logged but don't affect the loop exit code, matching the behavior of per-ticket hooks and preventing hook failures from masking successful loop execution

3. **Execution Timing**: Hooks run after the agent log summary is written but before cleanup/shutdown, ensuring they have complete loop information but can still perform actions like git pushes

4. **Variable Expansion**: Used the same pattern as per-ticket hooks with `os.Expand()` for consistency and to support both `$VAR` syntax and actual environment variables

## Invariants Compliance

- ✅ **Every behavior has a spec**: Added two scenarios to `specs/loop.feature`
- ✅ **Every spec has a test**: Created `testdata/loop/loop_on_complete_hook.txtar`
- ✅ **Spec before code**: Spec and test were part of the planned implementation
- ✅ **Zero external runtime dependencies**: Used only standard library packages (os/exec, strconv)
- ✅ **All tests pass**: `go test ./...` succeeds
- ✅ **Code compiles**: `go build ./...` succeeds

## Testing

- All existing tests continue to pass
- New test `TestLoop/loop_on_complete_hook` passes
- Full test suite: `go test ./...` ✅

## Future Considerations

The implementation provides a solid foundation for post-loop automation. Potential use cases include:
- Pushing changes after successful ticket processing
- Updating documentation after loops complete
- Sending notifications with loop statistics
- Conditional deployment based on loop results

No further work is required for this ticket.
