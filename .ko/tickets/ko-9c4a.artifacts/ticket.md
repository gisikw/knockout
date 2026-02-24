---
id: ko-9c4a
status: open
deps: []
created: 2026-02-24T23:13:23Z
type: task
priority: 2
---
# Add plan-questions to frontmatter schema and ko block --questions flag

Extend ticket frontmatter to support a `plan-questions` field — array of
question objects with id, question, context, and options.

`ko block <id> --questions '<json>'` writes questions into frontmatter and
sets status to blocked. Questions are optional (plain `ko block` still works
for unreasoned blocks).

## Schema

```yaml
plan-questions:
  - id: q1
    question: "Tabs or spaces?"
    context: "INVARIANTS.md says..."
    options:
      - label: "Spaces, 2-wide (Recommended)"
        value: spaces_2
        description: "Matches existing codebase convention"
      - label: "Tabs"
        value: tabs
        description: "Let the editor decide width"
```

Fields: `id` (required), `question` (required), `context` (optional),
`options` (required, array of `{label, value, description?}`).
