---
id: ko-b99d
status: open
deps: []
created: 2026-02-25T19:24:10Z
type: task
priority: 2
---
# Remove ko closed and ko resolved commands

ko closed and ko resolved are pure duplicates of 'ko ls --status=closed' and 'ko ls --status=resolved'. Remove both commands. Update help text and serve.go whitelist.
