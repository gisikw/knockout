Looking at the ticket:

**"ko ready should exclude snoozed tickets — if snooze date is in the future, ticket is not ready. Date-only values (e.g. 2026-05-01) are valid as of midnight on that date."**

This is describing **what to build**: specific behavior for the `ko ready` command that filters out tickets with future snooze dates. It includes clear specifications (date comparison logic, midnight handling for date-only values).

It's not asking "can we do this?" or reporting broken behavior—it's a feature request with a defined implementation requirement.

```json
{"disposition": "route", "workflow": "task"}
```
