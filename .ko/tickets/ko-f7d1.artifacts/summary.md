# Implementation Summary: Event Logging for Retries and Failures

## What Was Done

Successfully implemented comprehensive event logging for build step failures, retries, and build-level errors. The event log is now a complete record with no silent gaps.

### Core Changes

1. **New Event Types** (eventlog.go, buildhistory.go):
   - `node_fail`: Emitted when a node execution fails, includes error reason and attempt number
   - `node_retry`: Emitted before retrying a failed node, includes next attempt number
   - `build_error`: Emitted for build-level errors (hook failures), includes stage and reason

2. **Event Logger Threading** (build.go):
   - Added `log *EventLogger, hist *BuildHistoryLogger` parameters to `runNode()` function
   - Updated single call site in `runWorkflow()` to pass loggers through
   - Enabled retry loop to emit events during node execution

3. **Retry Loop Events** (build.go:257-284):
   - Emit `node_fail` on every failed attempt (before retry check)
   - Emit `node_retry` only when actually retrying (not on final failure)
   - Applied to both execution failures and disposition validation errors
   - Attempt numbers are 1-indexed for readability

4. **Hook Failure Events** (build.go:106, 113, 152-153):
   - Emit `build_error` for on_fail hook failures (2 call sites)
   - Emit `build_error` for on_succeed hook failures
   - Events capture which hook stage failed and the error message

### Test Coverage

Created/updated 3 test files:
- `build_retry_node_fail_events.txtar`: Node fails twice then succeeds - verifies node_fail and node_retry events
- `build_retry_exhausted.txtar`: Extended existing test to verify 3 node_fail events, 2 node_retry events
- `build_hook_failure_events.txtar`: on_succeed hook failure - verifies build_error event and correct build_complete sequencing

Added 3 behavioral scenarios to `specs/build_history.feature` documenting the new event types.

All pipeline tests pass (verified with `go test -run TestPipeline`).

## Notable Decisions

1. **Attempt Numbering**: Used 1-indexed attempt numbers (attempt+1) in events for user-facing readability, even though the internal retry loop is 0-indexed.

2. **Event Timing**: `node_fail` is emitted BEFORE the retry check, ensuring it appears even on the final attempt. `node_retry` is emitted ONLY when actually retrying (inside the retry condition).

3. **Hook Error Handling**: Hook failures now emit `build_error` but retain best-effort semantics - the error is logged but doesn't change the existing flow. on_succeed hook failures occur AFTER `build_complete("succeed")` is emitted, creating a correct sequence where the workflow succeeded but deployment failed.

4. **File Size**: eventlog.go grew from 125 to 157 lines (+32), buildhistory.go from 110 to 142 lines (+32) - both well under the 500-line limit per INVARIANTS.md.

## What to Know

- The event log is now a complete audit trail - every node execution produces either node_complete OR (node_fail → potentially node_retry → eventually node_complete)
- build_error events are additive information - they don't replace build_complete, they supplement it
- The 6-hour hung implement step scenario from the ticket would now show: node_start, node_fail (attempt 1, reason: "killed"), node_retry (attempt 2), node_start, node_complete (success on retry)
