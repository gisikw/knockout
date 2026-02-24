---
id: ko-f21c
status: open
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Support inline prompt text in pipeline nodes

Currently pipeline nodes require `prompt: filename.md` which loads from
`.ko/prompts/`. Support inline prompt text directly in the node config:

```yaml
- name: plan
  type: action
  prompt: |
    Review the ticket and produce a plan.
    Consider existing code patterns.
```

When `prompt:` value contains whitespace/newlines, treat it as inline text.
When it looks like a filename (no spaces, ends in .md), load from file as today.

Prereq for skills support in workflows (ko-4895).
