# Diagnosis: Missing Retry and Failure Events in Event Log

## Symptoms

When a build step fails and ko retries it, or when a step is killed (e.g., hung process), the retry/failure transition is invisible in the event log. The log shows `node_start` for the step, then nothing until the retry's `node_start` appears — or worse, nothing at all if the failure killed the build.

**Observed example:** A 6-hour hung implement step was killed externally. The event log showed `node_start` for implement with no corresponding complete/fail event. The loop retried successfully, but the retry was also invisible in the log.

**Observed in this ticket's history:**
```
{"event":"node_start","node":"classify",...}
{"event":"node_complete","node":"classify","result":"error",...}
```

This shows the final failure after all retries were exhausted, but provides no visibility into:
- Individual retry attempts (attempt 1, 2, 3...)
- What error occurred on each attempt
- When retries happened

## Root Cause

The retry logic in `runNode()` (build.go:227-263) silently retries failed node executions without emitting any events. The event emission happens at the workflow level in `runWorkflow()`, which only sees the final result from `runNode()`.

**Event emission points:**
1. `runWorkflow()` calls `log.NodeStart()` and `hist.NodeStart()` before calling `runNode()` (line 180-181)
2. `runNode()` executes with internal retry loop (lines 230-260) — **NO EVENTS EMITTED HERE**
3. `runWorkflow()` receives the final result and calls `log.NodeComplete()` and `hist.NodeComplete()` (lines 173-210)

**What happens during retry:**
```go
for attempt := 0; attempt < maxAttempts; attempt++ {
    // Execute node
    output, err = runPromptNode(...) // or runRunNode(...)

    if err != nil {
        if attempt+1 < maxAttempts {
            continue  // ← SILENT RETRY - no event emitted
        }
        return "", fmt.Errorf("failed after %d attempts: %v", maxAttempts, err)
    }

    // For decision nodes, validate disposition
    if node.Type == NodeDecision {
        if _, extractErr := extractDisposition(output); extractErr != nil {
            if attempt+1 < maxAttempts {
                continue  // ← SILENT RETRY - no event emitted
            }
            return "", fmt.Errorf("failed after %d attempts: %v", maxAttempts, extractErr)
        }
    }

    return output, nil
}
```

**Gap analysis:**
- When `attempt+1 < maxAttempts`, the code just `continue`s to the next retry
- No `node_fail` event is emitted for the failed attempt
- No `node_retry` event is emitted before the retry starts
- The only event that appears is when the workflow calls `NodeComplete()` with the final result

**Additional gaps:**
- Process kills/timeouts: If a node times out or is killed externally, the workflow may never complete, leaving orphaned `node_start` events with no corresponding `node_complete`
- Build-level errors: Hook failures (on_succeed, on_fail) don't emit dedicated `build_error` events, though they do call `BuildComplete` with "fail" outcome
- Hook failures aren't distinguished from node failures in the event log

## Affected Code

### Primary location: build.go:226-263 (`runNode()`)
This function contains the retry loop but has no access to the event loggers. It receives no `log` or `hist` parameters, so it cannot emit events even if we wanted it to.

### Secondary location: build.go:149-224 (`runWorkflow()`)
This function has the event loggers but doesn't know about retries. It delegates to `runNode()` and only sees the final outcome.

### Tertiary location: eventlog.go and buildhistory.go
Current event types:
- `node_start` ✓
- `node_complete` ✓
- `workflow_start` ✓
- `workflow_complete` ✓
- `build_start` ✓
- `build_complete` ✓
- `loop_ticket_start` ✓
- `loop_ticket_complete` ✓
- `loop_summary` ✓

**Missing event types:**
- `node_fail` (when a node fails before retry)
- `node_retry` (when a node is being retried)
- `build_error` (for build-level errors like hook failures)

## Recommended Fix

### 1. Pass event loggers to `runNode()`

Modify the `runNode()` signature to accept the event loggers:
```go
func runNode(ticketsDir string, t *Ticket, p *Pipeline, node *Node, model string,
    allowAll bool, timeout time.Duration, wsDir, artifactDir, wfName, histPath string,
    verbose bool, log *EventLogger, hist *BuildHistoryLogger) (string, error)
```

Update the call site in `runWorkflow()` (line 182) to pass the loggers.

### 2. Add new event emission methods

In `eventlog.go` and `buildhistory.go`, add:
```go
// NodeFail logs a failed node attempt with error details
func (l *EventLogger) NodeFail(ticket, workflow, node, reason string, attempt int) {
    l.emit(map[string]interface{}{
        "event":    "node_fail",
        "ticket":   ticket,
        "workflow": workflow,
        "node":     node,
        "reason":   reason,
        "attempt":  attempt,
    })
}

// NodeRetry logs a retry attempt
func (l *EventLogger) NodeRetry(ticket, workflow, node string, attempt int) {
    l.emit(map[string]interface{}{
        "event":    "node_retry",
        "ticket":   ticket,
        "workflow": workflow,
        "node":     node,
        "attempt":  attempt,
    })
}

// BuildError logs build-level errors (hook failures, etc.)
func (l *EventLogger) BuildError(ticket, stage, reason string) {
    l.emit(map[string]interface{}{
        "event":  "build_error",
        "ticket": ticket,
        "stage":  stage,
        "reason": reason,
    })
}
```

Mirror these methods in `buildhistory.go`.

### 3. Emit events in the retry loop

Modify `runNode()` to emit events:
```go
for attempt := 0; attempt < maxAttempts; attempt++ {
    var output string
    var err error

    if node.IsPromptNode() {
        output, err = runPromptNode(...)
    } else if node.IsRunNode() {
        output, err = runRunNode(...)
    } else {
        return "", fmt.Errorf("node '%s' has neither prompt nor run", node.Name)
    }

    if err != nil {
        // Emit node_fail event
        log.NodeFail(t.ID, wfName, node.Name, err.Error(), attempt+1)
        hist.NodeFail(t.ID, wfName, node.Name, err.Error(), attempt+1)

        if attempt+1 < maxAttempts {
            // Emit node_retry event
            log.NodeRetry(t.ID, wfName, node.Name, attempt+2)
            hist.NodeRetry(t.ID, wfName, node.Name, attempt+2)
            continue
        }
        return "", fmt.Errorf("node '%s' failed after %d attempts: %v", node.Name, maxAttempts, err)
    }

    // For decision nodes, validate disposition extraction
    if node.Type == NodeDecision {
        if _, extractErr := extractDisposition(output); extractErr != nil {
            // Emit node_fail event
            log.NodeFail(t.ID, wfName, node.Name, extractErr.Error(), attempt+1)
            hist.NodeFail(t.ID, wfName, node.Name, extractErr.Error(), attempt+1)

            if attempt+1 < maxAttempts {
                // Emit node_retry event
                log.NodeRetry(t.ID, wfName, node.Name, attempt+2)
                hist.NodeRetry(t.ID, wfName, node.Name, attempt+2)
                continue
            }
            return "", fmt.Errorf("node '%s' failed after %d attempts: %v", node.Name, maxAttempts, extractErr)
        }
    }

    return output, nil
}
```

### 4. Emit build_error events for hook failures

In `RunBuild()`, wrap hook execution calls:
```go
// Line 106 (on workflow error)
if err := runHooks(ticketsDir, t, p.OnFail, "", wsDir, hist.Path()); err != nil {
    log.BuildError(t.ID, "on_fail_hook", err.Error())
    hist.BuildError(t.ID, "on_fail_hook", err.Error())
}

// Line 135 (on_succeed hook failure)
if err := runHooks(ticketsDir, t, p.OnSucceed, changedFiles, wsDir, hist.Path()); err != nil {
    log.BuildError(t.ID, "on_succeed_hook", err.Error())
    hist.BuildError(t.ID, "on_succeed_hook", err.Error())
    AddNote(t, fmt.Sprintf("ko: on_succeed failed — %s", err.Error()))
    setStatus(ticketsDir, t, "blocked")
    // ... existing on_fail hook call
}
```

## Risk Assessment

### Low Risk
- **Adding new event types**: This is purely additive. Existing event consumers will ignore unknown event types.
- **Emitting events in retry loop**: This only adds information, doesn't change behavior.

### Medium Risk
- **Passing loggers to runNode()**: This changes the function signature. All call sites must be updated (currently only one call site in `runWorkflow()`).
- **Increased log volume**: Retries will now emit 2-4x more events. For a 3-retry failure:
  - Before: 2 events (`node_start`, `node_complete`)
  - After: 8 events (`node_start`, `node_fail`, `node_retry`, `node_fail`, `node_retry`, `node_fail`, `node_complete`)
  - Impact on disk space and log parsing tools should be considered.

### No Risk
- **Build history file**: Already append-only, additional events won't break existing parsers.
- **Event log file**: Truncated on each build, so no accumulation issues.

### Potential Side Effects
- **Log consumers**: Any external tools parsing the event log need to handle new event types gracefully (should already be doing this for forward compatibility).
- **Performance**: Minimal impact — event emission is fast (JSON marshal + append + sync), but sync is called more frequently. Not expected to be noticeable.

### What Could Go Wrong
1. **Incomplete implementation**: If we forget to update the call site, compilation will fail (good — fail fast).
2. **Event ordering**: Events are emitted synchronously, so ordering is guaranteed. No race condition risk.
3. **Nil pointer dereference**: If `log` or `hist` are nil, emit methods need nil checks (already present in current implementation).

## Edge Cases to Consider

1. **First attempt succeeds**: No `node_fail` or `node_retry` events are emitted. This is correct behavior.
2. **Timeout/kill during execution**: The context timeout in `runPromptNode()` or `runRunNode()` will return an error, which will be caught and logged as `node_fail`.
3. **Process killed externally**: If the process is killed with SIGKILL, no events can be emitted. The event log will end with `node_start`. This is unavoidable, but the build history shows this is happening (the ticket note mentioned a 6-hour hung process).
4. **Multiple workflows**: Each workflow execution tracks its own node execution. The `workflow` field in events distinguishes them.

## Summary

The root cause is straightforward: retry logic lives in `runNode()`, but event emission lives in `runWorkflow()`, and `runNode()` doesn't have access to the event loggers. The fix is equally straightforward: pass the loggers to `runNode()` and emit `node_fail` and `node_retry` events in the retry loop. Add `build_error` events for hook failures to complete the picture.

This is a clean fix with minimal risk — the only breaking change is a function signature update with a single call site.
