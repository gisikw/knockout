---
id: ko-3d1f
status: closed
deps: [ko-9c4a]
created: 2026-02-24T23:20:18Z
type: task
priority: 2
---
# Add ko questions command to read plan-questions as JSON

`ko questions <id>` reads the `plan-questions` field from ticket frontmatter
and outputs it as JSON to stdout.

```bash
ko questions ko-f65e
```

Output: the plan-questions array as JSON, or `[]` if no questions exist.

```json
[
  {
    "id": "q1",
    "question": "Should we keep backwards compatibility with pipeline.yml?",
    "context": "INVARIANTS.md says...",
    "options": [
      {"label": "Deprecation path (Recommended)", "value": "deprecate", "description": "..."},
      {"label": "Hard break", "value": "hard_break", "description": "..."}
    ]
  }
]
```

This is the read complement to `ko block --questions` (ko-9c4a) and
`ko answer` (ko-16e8). Punchlist server needs this to render question
UI for blocked tickets.

Depends on ko-9c4a (plan-questions schema must exist first).

## Notes

**2026-02-25 00:03:25 UTC:** ko: SUCCEED
