---
id: ko-c2a8
status: closed
deps: []
links: []
created: 2026-02-18T19:06:34Z
type: task
priority: 4
---
# Move .tickets/ to .ko/tickets/ — unify project state under .ko/

Currently ticket data lives in .tickets/ while pipeline config, builds, and prompts live in .ko/. Consolidating to .ko/tickets/ keeps all knockout state in one directory. Needs migration path: detect .tickets/ at old location, move or symlink. FindTicketsDir and all references need updating. This is a breaking change for existing projects — needs a migration on first access (move + symlink for backwards compat, or just move and update).
