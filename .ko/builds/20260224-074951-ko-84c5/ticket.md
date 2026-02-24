---
id: ko-84c5
status: open
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Support --json flag across ko commands

Add `--json` flag to ko commands that produce output. Enables scripting
and integration with other tools. `ko query` already does JSON; extend
the pattern to `ko show`, `ko ls`, `ko ready`, `ko blocked`, etc.
