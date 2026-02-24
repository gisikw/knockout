Looking at this ticket, it's asking for a specific code change:

- **What**: Rename the command `ko add-note` to `ko note`
- **Why**: Shorter, cleaner command
- **How**: Keep `add-note` as a hidden alias for backward compatibility

This is clearly a **task** - it describes exactly what to build (a renamed command with backward compatibility), not a problem to diagnose or a question to investigate.

```json
{"disposition": "route", "workflow": "task"}
```
