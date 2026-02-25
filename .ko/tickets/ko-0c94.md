---
id: ko-0c94
status: open
deps: [ko-6612]
created: 2026-02-25T23:29:21Z
type: task
priority: 2
---
# Remove ko triage command and subsidiaries

ko triage is superseded by ko update. Remove cmdTriage and all related code (--block, --questions, --answers flags on triage). Also remove ko blocked command if still present.
