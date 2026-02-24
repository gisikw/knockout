Looking at this ticket, I need to understand what type of work it describes:

The ticket states:
- **Problem**: specs/ directory has .feature files but they're not wired up to run
- **Proposed solutions**: 
  1. Introduce behave (Python BDD framework) and implement step definitions to make specs execute
  2. OR evaluate if txtar-based tests already cover the ground and .feature files are redundant

The core question is whether this is asking me to:
- **Build something** (implement behave integration + step definitions) — **task**
- **Investigate whether there's a problem** (are .feature files redundant?) — **research**
- **Fix something broken** — **bug**

The ticket is primarily asking me to evaluate a situation: do we need to wire up the .feature files, or are they already covered by existing tests and therefore unnecessary? This requires investigation and findings before committing to implementation.

The secondary path (if .feature files are needed) is a task—implementing behave integration.

This is fundamentally a **research** ticket: "Alternatively, evaluate whether the txtar-based tests already cover the same ground and the .feature files are redundant. If so, pick one approach and commit to it." The ticket is asking me to determine the right path forward before deciding what code work is needed.

```json
{"disposition": "route", "workflow": "research"}
```
