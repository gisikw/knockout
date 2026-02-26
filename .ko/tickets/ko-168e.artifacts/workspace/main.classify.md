Looking at this ticket, I can classify it straightforwardly:

**Analysis:**
- The ticket explicitly states what needs to be done: "Remove project/cwd from POST payload, use --project flag"
- It describes the implementation approach: the caller includes `--project` in the args array like any other flag
- The rationale is clear: simplifies the serve API and removes a special case
- There's a note that it's "Waiting for upstream changes to be merged" - this is a dependency, not ambiguity about what to build

This is clearly a **task** - it describes specific code changes to make (remove payload parameters, adjust how the endpoint processes requests) with a clear implementation path.

The note about waiting for upstream changes suggests this might be blocked, but the classification itself is unambiguous.

```json
{"disposition": "route", "workflow": "task"}
```
