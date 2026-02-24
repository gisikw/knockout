Perfect! I've created a comprehensive implementation plan. The plan:

## Summary

**Goal:** Make retry transitions and build failures visible in the event log by emitting `node_fail`, `node_retry`, and `build_error` events.

**Key findings from investigation:**
1. A detailed diagnosis already exists that analyzed the exact problem
2. Root cause: `runNode()` contains retry logic but has no access to event loggers
3. Solution is straightforward: thread loggers through, emit events in retry loop
4. Single call site to update, low risk additive change

**Plan structure:**
- **Tasks 1-2:** Add three new event emission methods to both event loggers
- **Task 3:** Update `runNode()` signature to accept loggers 
- **Tasks 4-6:** Emit events during retry loop (node_fail, node_retry)
- **Tasks 7-9:** Emit build_error for hook failures
- **Tasks 10-12:** Add three testscript tests covering retry events, exhausted retries, and hook failures
- **Task 13:** Update build_history.feature spec with new behavioral scenarios

The plan follows the existing patterns, stays within file size constraints, and has clear verification steps for each task.
