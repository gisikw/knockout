I'll analyze this ticket to classify the type of work it requires.

The ticket is asking for a feature to be implemented:
- **`ko blocked <id>`** — output the reason a specific ticket is blocked
- **`ko blocked`** — output reasons for all blocked tickets
- Consider: enforce that blocking reasons are captured in a structured way (frontmatter or a structured field) so they can be extracted programmatically

This ticket describes **what to build**: a new feature for the `ko` CLI to display blocking reasons for blocked tickets. It also mentions a consideration about how to structure the data (enforcing reasons in frontmatter/structured fields).

The expected output is:
1. Code changes to implement the `ko blocked` command
2. Potentially changes to how blocking reasons are stored/structured

This is clearly a **task** — it's asking for code changes and new functionality with clear requirements about what to build.

```json
{"disposition": "route", "workflow": "task"}
```
