Looking at this ticket, I need to classify the work required.

**Ticket Summary:**
- Title: "add-note should accept multiline text (heredoc/stdin)"
- This is a **feature request** that clearly describes what to build
- It specifies the desired behavior: the `add-note` command should accept multiline text via heredoc or stdin
- The ticket explicitly requests code changes to implement this capability

**Classification Reasoning:**
- The ticket describes **what to build** (multiline text input support), not **what's broken**
- There's a clear, specific implementation goal
- The previous attempt failed due to permission issues, but the requirement itself is well-defined
- This is not asking "can we do this?" (research) or reporting a bug - it's requesting a new feature with clear specifications

**Disposition:**
This should be routed to the `task` workflow.

```json
{"disposition": "route", "workflow": "task"}
```
