I'll analyze this ticket to classify the type of work it requires.

## Ticket Analysis

**Title:** "Support granular tool permissions in pipeline config"

**Key aspects:**
1. The ticket is requesting a feature to support granular tool permissions in pipeline config
2. There's a history showing that an architectural decision was made: "Decision: override, not merge. Node-level allowed_tools completely replaces parent lists."
3. The failure notes indicate that previous attempts hit an unresolved architectural decision, which has now been resolved
4. The work involves implementing this feature with the decided override semantics

**Classification reasoning:**

This ticket is asking for **code changes** to implement a specific feature (granular tool permissions in pipeline config). The expected output is:
- Code changes to the pipeline configuration system
- Implementation of the override semantics for `allowed_tools` at the node level
- Likely updates to configuration parsing, validation, and enforcement logic

The architectural decision has already been made (override vs merge semantics), removing the primary blocker. This is now ready to be implemented as a concrete feature.

This is a **task** because:
- It describes what to build (granular tool permissions with override semantics)
- The expected output is code changes
- The work is bounded and actionable given the architectural decision

```json
{"disposition": "route", "workflow": "task"}
```
