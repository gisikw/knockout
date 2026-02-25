Looking at this ticket, I need to classify the type of work required.

## Analysis

The ticket clearly specifies:
- **What to build**: A new `ko questions <id>` command
- **Expected behavior**: Read the `plan-questions` field from ticket frontmatter and output it as JSON
- **Example output**: A concrete JSON schema showing the expected structure
- **Context**: This is the read complement to existing `ko block --questions` and `ko answer` commands, with a dependency on ko-9c4a for the schema

This is a straightforward feature request with:
- Clear requirements (read plan-questions from frontmatter)
- Defined output format (JSON array with specific structure)
- Known dependency (ko-9c4a must exist first)
- No ambiguity about what needs to be done

The expected output is committed code implementing this command.

```json
{"disposition": "route", "workflow": "task"}
```
