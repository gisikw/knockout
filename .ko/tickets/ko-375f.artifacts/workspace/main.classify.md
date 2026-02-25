I'll analyze this ticket to determine the appropriate classification.

## Analysis

**Ticket Summary:**
"Migrate agent harnesses from YAML templates to shell scripts"

**Key Indicators:**

1. **Clear Scope**: The ticket explicitly describes *what to build* - a migration from YAML-based harnesses to shell scripts.

2. **Defined Architecture**: The notes indicate decisions have already been made:
   - Use executable wrapper approach
   - Namespace env vars with `KO_`
   - Replace (not maintain) YAML harnesses
   - Move binary fallback logic to shell
   - Pass prompt via `KO_PROMPT` environment variable

3. **Expected Output**: Code changes - converting existing YAML harnesses to shell scripts and removing YAML support.

4. **Not a Bug**: There's no assertion that something is broken or malfunctioning. This is a deliberate architectural change to ensure "agent invocation intent carries through, regardless of the actual agent harness."

5. **Not Research**: The fundamental questions have been answered (motivation, architecture, env var approach). The ticket is ready for implementation, not investigation.

**Classification: TASK**

This is clearly a code change request with defined requirements and architectural decisions already made. The work involves:
- Creating shell script harnesses
- Migrating built-in and user harnesses
- Removing YAML support
- Implementing the env var pattern

```json
{"disposition": "route", "workflow": "task"}
```
