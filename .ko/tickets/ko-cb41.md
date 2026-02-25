---
id: ko-cb41
status: open
deps: [ko-9c4a]
created: 2026-02-25T01:23:51Z
type: task
priority: 2
---
# Actionable node should emit structured plan-questions on fail

When the actionable node finds open questions in a plan and fails, it currently
emits a plain text reason string. The pipeline then calls applyFailOutcome which
writes it as a note and sets status to blocked.

For the punchlist plan-questions UI to work, the actionable node needs to produce
structured question data that lands in the ticket's plan-questions frontmatter
(the schema from ko-9c4a).

Options to consider:
- New disposition field: `{"disposition": "fail", "reason": "...", "plan_questions": [...]}`
  and teach applyFailOutcome to write plan-questions when present
- New disposition type: `{"disposition": "blocked", "plan_questions": [...]}`
- Have the actionable prompt extract questions into the schema format and the
  pipeline writes them via ko block --questions

The actionable prompt (actionable.md) needs updating to produce the right JSON
structure, and the pipeline code needs to route that structure into frontmatter.
