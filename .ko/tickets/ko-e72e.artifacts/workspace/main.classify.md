Looking at this ticket, I need to classify what kind of work it requires.

The ticket states:
- **What it's asking for**: A specific feature/improvement - when a ticket fails and gets retried, the agent should detect and reuse existing plan/artifacts rather than starting from scratch
- **Expected output**: Code changes to implement this detection and reuse logic
- **Nature of work**: Describes what to build (artifact reuse mechanism) not what's broken

This is clearly requesting a code change to add new functionality to the agent system. The work involves:
1. Detecting workspace/task files in artifact directories
2. Injecting them into prompt context
3. Modifying the implement node to reuse these artifacts on retry

This is a straightforward feature request with a clear implementation target.

```json
{"disposition": "route", "workflow": "task"}
```
