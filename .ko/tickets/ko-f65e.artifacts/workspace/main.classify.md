I'll analyze this ticket to determine the appropriate classification and workflow.

## Analysis

**Ticket Summary:**
The ticket proposes consolidating multiple `.ko/` top-level files (specifically `pipeline.yml` and potentially others like `agent.lock`, `agent.pid`, `agent.log`) into a single `config.yaml`. It describes desired functionality (a single place for pipeline config, agent harness overrides, and project-level settings) and notes this is a prerequisite for a separate piece of work (ko-1930).

**Key Indicators:**

1. **Describes what to build**: The ticket clearly articulates the desired end state - consolidating configuration into `config.yaml`
2. **Not asking "is this possible?"**: It's not a research question; it's a directive to implement a consolidation
3. **Not reporting broken behavior**: There's no assertion that current behavior is wrong or buggy
4. **Clear intended output**: Code changes (configuration restructuring and likely tooling updates to use the new structure)
5. **Prerequisite relationship**: It's framed as enabling work for ko-1930, suggesting it's a foundational change
6. **Medium discretion level**: Suggests some judgment is expected, but the core request is clear

This is clearly a **task** - it asks for specific code changes (consolidating configuration) with a well-defined outcome (single `config.yaml` file). While implementation details may require some investigation, the ticket itself is directive about what to build, not requesting investigation into whether something is possible or what's wrong.

```json
{"disposition": "route", "workflow": "task"}
```
