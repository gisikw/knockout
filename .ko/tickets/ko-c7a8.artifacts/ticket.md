---
id: ko-c7a8
status: open
deps: [ko-6612]
created: 2026-02-25T23:29:26Z
type: task
priority: 2
---
# Rewrite ko start/close/status as wrappers to ko update

ko start, ko close, and ko status should call the ko update function internally rather than having their own status-setting logic. e.g. ko start <id> becomes ko update <id> --status=in_progress.
