Looking at this ticket, I need to classify the work required.

The ticket is asking to:
1. **Implement** a new command (`ko update`)
2. **Define** its usage with specific flags and parameters
3. **Handle** auto-unblocking logic based on question/answer resolution

This is clearly asking for **code changes** to build a new feature. The expected output is committed code that implements the `ko update` command with all its specified functionality.

Even though the ticket mentions some logic that needs to be built (auto-unblocking), this isn't describing broken behavior that needs diagnosis—it's describing what the new feature should do when implemented.

```json
{"disposition": "route", "workflow": "task"}
```
