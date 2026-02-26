---
id: ko-e72e
status: open
deps: []
created: 2026-02-26T01:24:48Z
type: task
priority: 2
---
# Agent should resume from existing plan/artifacts on retry

When a ticket fails and gets retried, the agent starts from scratch. If a previous build produced a plan doc or other artifacts in .ko/tickets/<id>.artifacts/, the implement node should detect and reuse them rather than re-deriving everything. This would save significant time on retries — the triage and planning work is already done, the failure was likely in implementation or verification. Check for workspace/task.*.md files and inject them into the prompt context.
