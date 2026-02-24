I'll analyze this ticket to classify it.

## Analysis

**Ticket Summary:**
The ticket describes missing event logging when build steps fail and are retried. Specifically:
- When a step fails and ko retries it, no event is logged
- When a step is killed (e.g., hung process), the failure is not logged
- This makes diagnostics impossible from the event log alone

**What's described:**
- **Current state (problem):** Step failures and retries produce "silent gaps" in the event log
- **Expected state (solution):** Three new event types should be emitted: `node_fail`, `node_retry`, and `build_error`

**Key indicators:**
1. The ticket explicitly states "Expected events" with specific event types that should be emitted
2. It includes "Acceptance Criteria" that describe concrete implementation requirements
3. It's not asking "should we do this?" or "is this a problem?" — it's asserting the problem exists (with a concrete example: 6-hour hung implement step) and prescribing the solution
4. The expected output is implemented event logging in the codebase

**Classification reasoning:**
This falls between **task** and **bug**:
- It describes *what's broken* (missing event logs) → suggests **bug**
- It also describes *what to build* (new event types) → suggests **task**
- The presence of acceptance criteria and specific implementation details leans toward **task**
- However, the core issue is "events aren't being emitted when they should be" which is a malfunction

This is fundamentally a **bug** — the build system is not behaving correctly (missing critical diagnostics). The "fix" happens to involve adding code for missing events, but the classification should reflect that this is a functional defect being reported, not a feature request.

```json
{"disposition": "route", "workflow": "bug"}
```
