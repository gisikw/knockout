## Goal
Make build step failures, retries, and build-level errors visible in the event log by emitting node_fail, node_retry, and build_error events.

## Context

**Event Logging Infrastructure:**
- `eventlog.go` (125 lines): Per-build event logger, truncated on each build
- `buildhistory.go` (110 lines): Per-ticket append-only JSONL logger
- Both use identical event emission pattern via `emit()` method

**Current Event Flow:**
- `runWorkflow()` (build.go:159-233) emits `node_start` before calling `runNode()`
- `runNode()` (build.go:236-272) contains retry loop with `max_retries` attempts
- `runWorkflow()` emits `node_complete` after `runNode()` returns
- Gap: retry loop in `runNode()` has no access to event loggers, so retries are invisible

**Retry Mechanisms:**
- Node execution failures (timeout, non-zero exit, process killed)
- Invalid disposition JSON on decision nodes
- Both retry up to `max_retries + 1` total attempts (default: 3 attempts)

**Build-Level Errors:**
- Hook failures in `runHooks()` (build.go:618-652)
- Called from: on_fail hooks, on_succeed hooks, on_close hooks
- Hook failures currently logged via `BuildComplete("fail")` but no dedicated event

**Existing Event Types:**
- build_start, build_complete, workflow_start, workflow_complete
- node_start, node_complete
- loop_ticket_start, loop_ticket_complete, loop_summary

**Testing Pattern:**
- Specs in `specs/*.feature` (gherkin, not executable)
- Tests in `testdata/pipeline/*.txtar` using testscript
- Test helpers in `build_test.go` for running ko commands in test environments

**Prior Diagnosis:**
A detailed diagnosis document already exists at `.ko/tickets/ko-f7d1.artifacts/diagnosis.md` with recommended implementation approach. The diagnosis confirms:
- Root cause: `runNode()` lacks event logger parameters
- Solution: Pass loggers to `runNode()`, emit events in retry loop
- Risk: Low (additive changes, single call site update)

## Approach

Add three new event types (node_fail, node_retry, build_error) to both event loggers. Thread the event loggers through to `runNode()` so it can emit events during its retry loop. Emit node_fail when an attempt fails, node_retry before retrying, and build_error for hook failures. The implementation follows the existing event emission pattern and has a single call site to update.

## Tasks

1. [eventlog.go:NodeFail, NodeRetry, BuildError methods] — Add three new event emission methods: NodeFail(ticket, workflow, node, reason, attempt), NodeRetry(ticket, workflow, node, attempt), BuildError(ticket, stage, reason). Follow the existing pattern of NodeStart/NodeComplete. Each method calls emit() with appropriate event fields. NodeFail includes attempt number and error reason. NodeRetry includes next attempt number. BuildError includes stage (e.g., "on_succeed_hook") and reason.
   Verify: Code compiles, eventlog.go stays under 200 lines (currently 125).

2. [buildhistory.go:NodeFail, NodeRetry, BuildError methods] — Mirror the three new methods from eventlog.go. Identical signatures and field structure, just emitting to the build history file instead.
   Verify: Code compiles, buildhistory.go stays under 150 lines (currently 110).

3. [build.go:236 runNode signature] — Add `log *EventLogger, hist *BuildHistoryLogger` parameters to runNode function signature. Update the call site at build.go:191 in runWorkflow() to pass the loggers.
   Verify: Code compiles, no other call sites exist (grep confirms single call site).

4. [build.go:251-256 retry loop emit node_fail] — In the first error branch of the retry loop (execution failures), emit node_fail before the retry check. Call log.NodeFail() and hist.NodeFail() with the error message and current attempt number (attempt+1 for 1-indexed). Place emission before the "continue if more retries" logic.
   Verify: Logic emits fail event even on final attempt before returning error.

5. [build.go:251-256 retry loop emit node_retry] — In the first error branch, emit node_retry inside the "attempt+1 < maxAttempts" block (just before continue). Call log.NodeRetry() and hist.NodeRetry() with next attempt number (attempt+2).
   Verify: Retry event only emitted when actually retrying (not on final failure).

6. [build.go:260-266 disposition validation emit node_fail and node_retry] — In the decision node disposition validation error branch, mirror the event emissions from task 4 and 5. Emit node_fail with extractErr.Error() and attempt number, then emit node_retry before continue if more attempts remain.
   Verify: Both error paths in retry loop emit symmetric events.

7. [build.go:106 on_fail hook] — After the existing runHooks() call for on_fail (line 106), emit build_error if the hook fails. Wrap the call: `if err := runHooks(...); err != nil { log.BuildError(t.ID, "on_fail_hook", err.Error()); hist.BuildError(t.ID, "on_fail_hook", err.Error()) }`. The hook is best-effort so don't change error handling.
   Verify: Hook failure emits build_error but doesn't break existing flow.

8. [build.go:113 workflow on_fail hook] — Mirror task 7 for the on_fail hook at line 113.
   Verify: Both on_fail hook sites emit build_error consistently.

9. [build.go:143-148 on_succeed hook failure] — The on_succeed hook failure already has error handling. After line 144 where it emits the note and sets blocked status, add build_error emissions: `log.BuildError(t.ID, "on_succeed_hook", err.Error()); hist.BuildError(t.ID, "on_succeed_hook", err.Error())`. Place after the AddNote call, before the runHooks on_fail call.
   Verify: Hook failure path emits build_error before calling on_fail hooks.

10. [testdata/pipeline/build_retry_node_fail_events.txtar] — Create new test: node that fails twice then succeeds. Pipeline with max_retries: 2. Fake LLM that fails on first two invocations (using attempt counter file), succeeds on third. After build, verify JSONL contains: 2 node_fail events (attempt 1, attempt 2), 2 node_retry events (attempt 2, attempt 3), final node_complete with result "done". Check event ordering and attempt numbers.
   Verify: `go test -run TestScript/pipeline/build_retry_node_fail_events` passes.

11. [testdata/pipeline/build_retry_exhausted_events.txtar] — Extend existing build_retry_exhausted.txtar test. After build fails, add assertions for JSONL: 3 node_fail events (attempts 1, 2, 3), 2 node_retry events (attempts 2, 3), final node_complete with result "error". Verify no retry event after final failure.
   Verify: `go test -run TestScript/pipeline/build_retry_exhausted` passes.

12. [testdata/pipeline/build_hook_failure_events.txtar] — Create new test: on_succeed hook that fails. Pipeline with on_succeed hook that exits 1. After build, verify JSONL contains build_error event with stage "on_succeed_hook" and ticket ends in blocked status. Verify workflow_complete is "succeed" (since workflow finished) but BuildComplete is "fail" (since hook failed).
   Verify: `go test -run TestScript/pipeline/build_hook_failure_events` passes.

13. [specs/build_history.feature:51-70] — Add three new scenarios to the build history spec: "Retry attempts emit node_fail and node_retry events" (node fails twice, verify 2 node_fail + 2 node_retry), "Exhausted retries emit node_fail events" (verify node_fail on each attempt, final node_complete is error), "Hook failures emit build_error events" (on_succeed fails, verify build_error event). Keep scenarios behavioral, not implementation-focused.
   Verify: Scenarios describe observable behavior from JSONL parsing.

## Open Questions

None — the diagnosis document provides comprehensive implementation guidance with minimal ambiguity. The change is additive (new events), has a single call site to update, and follows existing event emission patterns. The event schema (field names, structure) matches the diagnosis recommendations which align with existing event types.
