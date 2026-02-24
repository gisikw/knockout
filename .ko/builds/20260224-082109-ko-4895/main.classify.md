I'll analyze this ticket to classify it and determine the appropriate workflow.

**Ticket Analysis:**

The ticket "Support skills in pipeline workflows" is requesting a feature to allow pipeline nodes to invoke skills. Let me break down what's being asked:

1. **Clear objective**: Enable skills to be invoked in pipeline workflows, with three levels of progression (making skills available, using skill names directly, and inline prompt references)

2. **What's being built**: A feature that doesn't currently exist - the ability for pipeline nodes to call skills rather than just raw prompts

3. **Open question noted**: The ticket explicitly acknowledges a design decision about how to handle the Claude Code adapter constraint (no `--add-skills-dir` flag), with proposed options

4. **Expected output**: Code changes to implement this feature, plus decisions on the configuration surface

**Classification:**

This is a **task** ticket. While it includes an open question about implementation approach, the core request is clearly "build this feature" - specifically, add support for skills in pipeline workflows. The ticket describes *what to build* (skill invocation support) with clear progression steps, not *what's broken* that needs diagnosis.

The open question about the adapter approach is an architectural consideration that should be resolved during implementation, not a blocker preventing classification.

```json
{"disposition": "route", "workflow": "task"}
```
