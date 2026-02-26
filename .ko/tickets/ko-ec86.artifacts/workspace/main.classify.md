Looking at this ticket, I need to classify what type of work it represents.

**Analysis:**

The ticket states: "Support --all flag on `ko ls` which induces it to include closed tickets (this is for parity with the ko serve SSE behavior)"

This is clearly describing **what to build** — a new feature (adding an `--all` flag to an existing command). The ticket:
- Specifies the exact feature needed (a `--all` flag)
- Explains its purpose (include closed tickets)
- Provides context (parity with existing behavior)
- Has a clear, defined fix

This is not asking questions or requesting investigation. It's not reporting unexpected/broken behavior. It's a straightforward feature request with clear requirements.

```json
{"disposition": "route", "workflow": "task"}
```
