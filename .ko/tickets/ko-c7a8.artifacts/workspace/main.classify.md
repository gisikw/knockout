Looking at this ticket, it's asking for a specific refactoring task:

**What it describes**: Rewrite three commands (ko start, ko close, ko status) to use ko update internally rather than having duplicate status-setting logic.

**Key characteristics**:
- Clear scope: three specific commands to refactor
- Known solution approach: make them wrappers around `ko update` with specific `--status` flags
- Expected output: refactored code that maintains existing behavior
- No diagnosis needed: not reporting a problem, but proposing a code organization improvement

This is a straightforward refactoring task with a well-defined implementation path.

```json
{"disposition": "route", "workflow": "task"}
```
