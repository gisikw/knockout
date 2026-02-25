Looking at this ticket, I need to classify what kind of work it represents.

## Analysis

The ticket describes:
1. **Current state**: Multiple top-level files in `.ko/` (pipeline.yml, agent.lock, agent.pid, agent.log, prompts/)
2. **Desired state**: Consolidate into a single `.ko/config.yaml`
3. **Rationale**: Single place for pipeline config, agent harness overrides, and project-level settings
4. **Context**: Prerequisite for agent harness extraction work (ko-1930)

The ticket is clearly **prescriptive** — it describes what to build, not what's broken. It specifies:
- What files should be consolidated
- What the output should contain
- Why this consolidation is valuable

This is not:
- A **bug**: There's no broken behavior being reported
- A **research** question: We're not investigating "should we?" or "can we?" — the ticket asserts this is the right approach

This is a **task**: It asks for code changes (consolidating configuration into a single file) with a clear, specified outcome.

```json
{"disposition": "route", "workflow": "task"}
```
