Looking at this ticket, I need to classify the type of work it represents.

**Ticket Analysis:**
- "Rename ko create to ko add" - This is a clear directive to change existing functionality
- The main command should change from `ko create` to `ko add`
- `create` should be kept as a hidden alias for backward compatibility
- This is requesting specific code changes with a known solution path

**Classification:**
This is a **task** ticket. It:
1. Describes what to build (rename the command)
2. Specifies the expected output (code changes with backward compat alias)
3. Has a clear, defined fix
4. Is not reporting broken behavior
5. Is not asking for investigation

```json
{"disposition": "route", "workflow": "task"}
```
