Looking at this ticket:

**"Ensure all output commands support --json flag"**

The ticket:
- **Describes what to build**: A feature requirement to add `--json` flag support to various commands
- **Provides clear scope**: Lists specific commands that need it (show, ls, ready, blocked, triage, agent status) and notes which ones already have it
- **Specifies the expected output**: Committed code changes adding the flag support
- **Is not reporting broken behavior**: It's not saying these commands are broken, just that they should support an additional flag for machine-readable output

This is clearly a **task** - it's a feature implementation with a well-defined requirement.

```json
{"disposition": "route", "workflow": "task"}
```
