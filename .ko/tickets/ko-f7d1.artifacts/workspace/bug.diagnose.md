## Summary

I've completed the diagnosis and written it to `diagnosis.md`. Here's what I found:

**Root Cause:** The retry logic in `runNode()` silently retries failed executions without emitting any events. Event emission happens at the workflow level in `runWorkflow()`, which only sees the final result — not the individual retry attempts.

**Key Problem:** The `runNode()` function doesn't receive the event logger parameters (`log` and `hist`), so it *cannot* emit events even if we wanted it to. When a retry happens (lines 243-244, 252-253 in build.go), the code just `continue`s to the next attempt with no visibility.

**What's Missing:**
1. `node_fail` events — when a step fails before retry (with error reason and attempt number)
2. `node_retry` events — when a failed step is being retried (with next attempt number)
3. `build_error` events — when build-level errors occur (hook failures, etc.)

**The Fix:** Pass the event loggers to `runNode()`, add the three new event types to `eventlog.go` and `buildhistory.go`, and emit them in the retry loop. This is a clean, low-risk change with only one function signature to update.

The diagnosis includes specific code changes, risk assessment, and edge case analysis.
