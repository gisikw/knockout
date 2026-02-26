I'll classify this ticket by reading its content carefully.

The ticket clearly states:
- **What to build**: "Add support for `require_clean_tree: true` in pipeline.yml"
- **Desired behavior**: A configuration option that prevents builds when uncommitted changes exist (outside `.ko/`)
- **Specification details**: The ticket includes notes clarifying the interpretation (prevent if dirty, not if clean)

The ticket describes **what to implement** (a new feature with clear requirements), not what's broken. It provides:
- A specific configuration option name
- Expected behavior (validation logic)
- Implementation notes in the prior context with a detailed plan
- Clear technical approach

This is a **task** — it's a feature implementation request with a complete specification.

```json
{"disposition": "route", "workflow": "task"}
```
