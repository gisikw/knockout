---
id: ko-b10a
status: closed
deps: [ko-8a7c]
links: []
created: 2026-02-24T03:04:21Z
type: task
priority: 2
---
# Example pipeline: classify → task/research/bug workflow routing

Replace `examples/default/` with a new example pipeline that routes tickets
through classify (haiku) into one of three workflows:

```yaml
workflows:
  main:
    - name: classify
      type: decision
      prompt: classify.md
      model: claude-haiku-4-5-20251001
      routes: [task, research, bug]

  task:
    - name: plan
      type: action
      prompt: plan.md
    - name: actionable
      type: decision
      prompt: actionable.md
    - name: implement
      type: action
      prompt: implement.md
    - name: verify
      type: action
      run: just test
    - name: review
      type: decision
      prompt: review.md

  research:
    - name: investigate
      type: action
      prompt: investigate.md  # close with findings

  bug:
    - name: diagnose
      type: action
      prompt: diagnose.md
    - name: assess
      type: decision
      prompt: assess.md  # fixable? → route to task | close with findings
      routes: [task]
```

Classify is cheap (haiku) and answers: is this a task (produce code), research
(produce answers), or bug (diagnose first, then maybe code)?

Prompts for each node need writing. Depends on artifact directory (ko-8a7c)
being available so nodes can share codebase context.
