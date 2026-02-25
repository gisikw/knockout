Looking at this ticket, I need to classify what kind of work it requires.

The ticket is asking to:
1. Remove the `ko query` command
2. Remove its tests
3. Remove the serve whitelist entry

The rationale is clear: `ko query` is redundant because other list commands (`ls`, `ready`, `blocked`, `closed`, `resolved`) now support `--json`, collectively covering all statuses.

This is a straightforward **task**. The ticket:
- Describes specific code changes to make (remove command, tests, whitelist entry)
- Provides clear reasoning for why the change is safe (no breakage expected since `ko serve` uses `ListTickets()` directly)
- Expects committed code as the output
- Is not reporting unexpected behavior or asking for investigation

```json
{"disposition": "route", "workflow": "task"}
```
