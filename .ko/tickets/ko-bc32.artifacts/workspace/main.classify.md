Based on the ticket description, this is clearly a **task**.

The ticket explicitly describes what needs to be built:
- A new `ko block` command that acts as a convenience wrapper
- It should accept parameters: `<id>`, optional `[reason]`, and optional `[--questions json]`
- It should translate to `ko update <id> --status=blocked` with appropriate flags

This is a straightforward feature request with a clear specification of the expected behavior and output. The implementation involves adding a new command to the codebase.

```json
{"disposition": "route", "workflow": "task"}
```
