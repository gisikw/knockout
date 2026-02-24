## Analysis

This ticket is requesting the implementation of a new command `ko answer` with specific behavior:

- Takes a ticket ID and JSON mapping of question IDs to answers
- Partially answers questions (only those in the JSON are resolved)
- Removes answered questions from frontmatter
- Appends timestamped notes for each answer
- Clears the `plan-questions` field and sets status to `open` when all questions are answered
- Keeps status as `blocked` if questions remain

This is a **feature request** with:
- Clear, specific requirements for what needs to be built
- Defined input format (`ko answer <id> '<json>'`)
- Defined output behavior (frontmatter modifications, note appending, status changes)
- No assertion of existing broken behavior
- No investigation needed — the specification is complete

This is straightforward implementation work: add a new command to the `ko` CLI that parses arguments, reads ticket files, modifies frontmatter, and appends notes.

```json
{"disposition": "route", "workflow": "task"}
```
