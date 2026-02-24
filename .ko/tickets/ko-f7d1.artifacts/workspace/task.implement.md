Perfect! All tests pass. Let me create a summary of what was implemented:

## Summary

I successfully implemented retry and build failure event logging for the knockout build system. All 13 tasks from the plan were completed:

### Code Changes

**1-2. Event Logger Methods (eventlog.go, buildhistory.go)**
- Added `NodeFail(ticket, workflow, node, reason, attempt)` - logs node execution failures with error reason and 1-indexed attempt number
- Added `NodeRetry(ticket, workflow, node, attempt)` - logs retry events with next attempt number
- Added `BuildError(ticket, stage, reason)` - logs build-level errors like hook failures
- Both files stay well under their line limits (eventlog.go: 157/200, buildhistory.go: 142/150)

**3. Threading Event Loggers (build.go:191, 236)**
- Updated `runNode` signature to accept `log *EventLogger` and `hist *BuildHistoryLogger` parameters
- Updated the single call site in `runWorkflow()` to pass the loggers

**4-6. Retry Loop Event Emissions (build.go:251-273)**
- Emit `node_fail` before retry check for both execution failures and disposition validation errors
- Emit `node_retry` only when actually retrying (not after final failure)
- Both error paths in the retry loop now emit symmetric events

**7-9. Hook Failure Events (build.go:106, 113, 151)**
- Wrapped `runHooks()` calls to emit `build_error` when on_fail hooks fail
- Added `build_error` emission when on_succeed hooks fail (after AddNote, before on_fail hooks)
- Hook failures are now visible in the event log

### Tests

**10. build_retry_node_fail_events.txtar**
- Tests a node that fails twice then succeeds on the third attempt
- Verifies 2 node_fail events, 2 node_retry events, and final node_complete with result "done"

**11. build_retry_exhausted.txtar (extended)**
- Added assertions for 3 node_fail events, 2 node_retry events, and node_complete with result "error"
- Confirms no retry event after final failure

**12. build_hook_failure_events.txtar**
- Tests on_succeed hook failure
- Verifies build_error event with stage "on_succeed_hook"
- Confirms build_complete shows "succeed" (workflow finished before hook failed)

**13. specs/build_history.feature**
- Added 3 behavioral scenarios describing the new event types
- Scenarios cover retry attempts, exhausted retries, and hook failures

### Verification

- All code compiles successfully
- All new tests pass
- All existing tests continue to pass
- The event log now provides a complete record with no silent gaps between events

The implementation followed the existing event emission pattern, had minimal risk (single call site update), and makes build failures and retries fully observable in the event log.
