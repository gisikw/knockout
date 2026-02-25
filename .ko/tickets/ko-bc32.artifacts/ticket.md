---
id: ko-bc32
status: open
deps: [ko-6612]
created: 2026-02-25T23:29:31Z
type: task
priority: 2
---
# Add ko block as wrapper to ko update

ko block <id> [reason] [--questions json]

Convenience wrapper: ko block <id> reason -> ko update <id> --status=blocked -d reason
ko block <id> --questions json -> ko update <id> --status=blocked --questions json
