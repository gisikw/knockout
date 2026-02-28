Looking at this ticket:

"Add triage field to ticket frontmatter — a free-text string (e.g. 'unblock this ticket', 'break this apart'). Settable via ko add and ko update."

This is a **task**. It describes specific functionality to build: a new `triage` field in ticket frontmatter, settable via existing commands. The expected output is code changes.

```json
{"disposition": "route", "workflow": "task"}
```
