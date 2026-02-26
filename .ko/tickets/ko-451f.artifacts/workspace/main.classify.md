Reading this ticket, I can identify the following:

1. **Precondition**: "After ko update is implemented and triage/subsidiaries removed" - this is a dependency that needs to be completed first
2. **Clear actions**: 
   - Update README to reflect new command surface
   - Run `just install`
3. **Expected output**: Updated documentation and successful installation run

This ticket describes **what to build** (update documentation, run installation) rather than what's broken. It's a follow-up task that depends on another feature being completed. The actions are straightforward and the expected outcome is committed code changes (README updates).

```json
{"disposition": "route", "workflow": "task"}
```
