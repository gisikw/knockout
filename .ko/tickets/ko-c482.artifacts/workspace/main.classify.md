Looking at the ticket, this is a straightforward feature request:

- **What**: Add an `actionable` field to `ko agent status --json` output
- **Condition**: Field should be true when ready queue is non-empty OR triageable tickets exist
- **Purpose**: Consumers (XO, Punchlist) use it to gate showing the agent toggle
- **Expected output**: Code changes to implement this field

This describes *what to build*, not *what's broken*, with clear requirements.

```json
{"disposition": "route", "workflow": "task"}
```
