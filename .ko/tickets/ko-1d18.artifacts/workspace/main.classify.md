Looking at this ticket:

**"Add ko snooze <id> <date> as shorthand for ko update <id> --set snooze=<date>"**

This is clearly asking for a new feature—a shorthand command for setting the snooze field on a ticket. It describes what to build, not what's broken. The implementation is straightforward: add a new `ko snooze` subcommand that internally calls the existing `ko update` command with the snooze parameter.

```json
{"disposition": "route", "workflow": "task"}
```
