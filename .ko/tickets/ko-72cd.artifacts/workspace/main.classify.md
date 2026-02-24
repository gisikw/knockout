I'm reading this ticket carefully.

## Analysis

**What the ticket asks for:**
- "Once `resolved` status exists (ko-dd51), the pipeline needs a way to land tickets there instead of `closed` on success."
- This is framed as a feature request with clear direction and multiple proposed solutions.

**The options presented:**
1. New disposition type
2. Workflow-level config (`on_success: resolved` vs `on_success: closed`)
3. Convention-based defaults with per-node overrides

**Key indicator:**
The ticket explicitly states "The right answer probably involves both — a yaml-level default per workflow, plus a disposition that lets individual decision nodes override." This shows the work is **what to build**, not **what's broken** or **whether something is a problem**.

The ticket describes the desired functionality (pipeline support for resolved status as an outcome), proposes architectural approaches, and indicates implementation should involve both configuration and code changes.

This is a **task**: it's asking for code changes to implement a feature (workflow configuration support and disposition types), with clear requirements and proposed solutions.

```json
{"disposition": "route", "workflow": "task"}
```
