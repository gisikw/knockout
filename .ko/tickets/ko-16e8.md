---
id: ko-16e8
status: open
deps: [ko-9c4a]
created: 2026-02-24T23:13:24Z
type: task
priority: 2
---
# Add ko answer command for submitting plan question answers

`ko answer <id> '<json>'` where json maps question IDs to answer strings:

```bash
ko answer ko-f65e '{"q1":"Spaces, 2-wide","q2":"I will fix manually"}'
```

Behavior:
- Partial answers supported — only resolves the questions whose IDs appear in the JSON
- Each answered question is removed from `plan-questions` in frontmatter
- Each answer appended as a timestamped note: `Plan answer (q1): Tabs or spaces? → Spaces, 2-wide`
- When last question is answered, `plan-questions` field is cleared entirely and status set to `open`
- If questions remain, status stays `blocked`
