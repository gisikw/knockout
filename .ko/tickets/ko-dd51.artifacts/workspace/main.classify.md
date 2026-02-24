I'll analyze this ticket to classify it.

## Analysis

The ticket is titled "Add a `resolved` status between `in_progress` and `closed`."

**What the ticket describes:**
- A clear request to add a new status value to the system
- Specific placement: between `in_progress` and `closed`
- Defined use cases (research questions, bug wontfixes, completed work needing sign-off)
- Expected behavioral changes:
  - `ko ready` should not surface resolved tickets
  - `ko resolved` or `ko ls --status resolved` should list them
  - Agent loop should transition to `resolved` instead of `closed` for certain workflows

**Classification reasoning:**

This is asking for **what to build** rather than reporting **what's broken**:
- It's a feature request for a new status state
- The desired behavior is clearly specified
- The implementation path is clear: add a status value, update CLI commands to handle it, update agent transition logic
- The expected output is committed code that implements this status system

This is not:
- A bug (no broken behavior reported)
- Research (no investigation needed; the requirement is fully specified)

**Disposition:** This is a **task** — implementing a feature with clear requirements and expected code changes.

```json
{"disposition": "route", "workflow": "task"}
```
