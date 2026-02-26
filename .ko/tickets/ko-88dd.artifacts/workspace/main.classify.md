Looking at this ticket, I can identify the key aspects:

**What the ticket describes:**
- A specific feature request: "Add support for `require_clean_tree: true` in pipeline.yml"
- Clear intent: "prevent an agent from starting if the working tree is clean (ignoring ./.ko)"
- This describes what to build, not what's broken

**Classification indicators:**
- The ticket specifies exactly what needs to be implemented (a new configuration option)
- It explains the desired behavior (validation/prevention logic)
- There's no indication of current broken behavior or unexpected issues
- The work involves adding a feature to existing code

This is a **task** - it's a feature implementation with a clear specification of what needs to be built.

```json
{"disposition": "route", "workflow": "task"}
```
