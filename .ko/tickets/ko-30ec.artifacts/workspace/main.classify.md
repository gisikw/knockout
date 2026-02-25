Looking at this ticket, I can clearly identify what's being requested:

**What the ticket is asking for:**
- Add an optional "project" field to the JSON body of the POST /ko endpoint
- Implement project resolution logic (via registry for #tags, or as absolute path)
- Set the working directory for exec based on the resolved project
- If the field is absent, fall back to using cwd
- This mirrors the pattern already used in the subscribe endpoint

**Key indicators this is a task:**
1. Explicitly describes *what to build* - a new feature (project field support)
2. Provides concrete implementation details (field name, behavior, resolution pattern)
3. References an existing pattern to follow (subscribe endpoint)
4. Has clear acceptance criteria

This is straightforward feature development with a known pattern to follow.

```json
{"disposition": "route", "workflow": "task"}
```
