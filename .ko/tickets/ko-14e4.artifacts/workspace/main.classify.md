Looking at the ticket title:

**"ko agent start should run ko agent triage on all triageable tickets before starting work on ready tickets"**

This is a clear specification of desired behavior. It describes *what to build* — a feature enhancement to the `ko agent start` command to add a triage step before processing ready tickets.

This is not:
- A question or investigation request (research)
- An assertion of broken behavior (bug)

It's an implementation request with a clear definition of what should happen.

```json
{"disposition": "route", "workflow": "task"}
```
