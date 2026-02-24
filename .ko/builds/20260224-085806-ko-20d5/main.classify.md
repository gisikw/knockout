I'll analyze this ticket to classify it appropriately.

## Analysis

**Ticket Type: Task**

The ticket clearly describes **what to build**, not what's broken:
- It specifies a concrete feature to implement: `step_timeout` configuration
- It provides a detailed design with YAML structure examples
- It lists explicit acceptance criteria for the implementation
- The expected output is committed code changes

**Key indicators this is a task:**
1. **Clear requirements**: Add timeout configuration at pipeline and per-node levels
2. **Design provided**: Specific YAML structure, default value (15m), implementation approach (`exec.CommandContext`)
3. **Acceptance criteria**: Measurable, testable outcomes
4. **Known solution path**: The problem (hangs) and solution (timeouts) are well-defined

While the ticket mentions a real problem from production (6-hour hang in exo-6de5), the focus is not on diagnosing an unknown issue—it's on implementing a specific feature to prevent future occurrences.

```json
{"disposition": "route", "workflow": "task"}
```
