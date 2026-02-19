---
id: ko-3ff7
status: closed
deps: [ko-fb9e]
links: []
created: 2026-02-19T02:22:44Z
type: task
priority: 2
---
# Include modified timestamp in ko query JSONL output

## Notes

**2026-02-19 02:22:49 UTC:** ko query currently outputs JSONL with id, title, status, type, priority, deps, links, created, assignee. Add a 'modified' field (RFC3339 timestamp from file mtime) so consumers like punchlist-server can sort by mtime without statting ticket files themselves. This is the field punchlist-server will use for display ordering.
