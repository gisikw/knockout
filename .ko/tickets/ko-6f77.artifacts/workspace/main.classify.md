Looking at this ticket, I can see it's describing specific work to be done:

**What it asks for:**
1. Add `--project` flag to specific commands (add, ls, ready, agent loop, agent init, agent start, agent stop, agent status)
2. Strip hashtag auto-detection from the title in `ko add` — use `--project` exclusively
3. Implementation detail: modify `resolveProjectTicketsDir()` to accept an optional `--project` flag value

This is a clear **feature/refactoring request** with defined changes. It specifies:
- Which commands need the flag
- What behavior to change (remove hashtag parsing, require explicit flag)
- How to implement it (modify a specific function)

The expected output is committed code implementing these changes.

```json
{"disposition": "route", "workflow": "task"}
```
