Looking at this ticket, I need to determine what type of work it describes.

The ticket clearly states what needs to be built:
- Extend ticket frontmatter to support a `plan-questions` field with a specific schema (array of question objects with id, question, context, and options)
- Add a `ko block <id> --questions '<json>'` flag that writes questions into frontmatter and sets status to blocked
- Questions should be optional (plain `ko block` still works without them)

The ticket includes:
- Detailed schema specification showing the exact structure needed
- Clear field requirements (id and question required, context optional, options required with specific shape)
- Specific behavior expectations (status is set to blocked, questions are optional)

This is unambiguously describing **what to build**, not describing broken behavior or asking for investigation. The implementation is straightforward: extend the frontmatter schema validation, add the CLI flag, and update the blocking logic to handle the new questions field.

```json
{"disposition": "route", "workflow": "task"}
```
