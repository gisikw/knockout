---
id: ko-57d3
status: open
deps: []
created: 2026-02-25T05:50:03Z
type: task
priority: 2
---
# ko ls: add --limit and --json flags

Add two flags to `ko ls` (and ideally ready, blocked, resolved, closed too):

- `--limit N` — return only the first N items after sorting
- `--json` — output as JSON array instead of human-readable table

The --json flag should emit the same shape as `ko query` but as a JSON array
(not JSONL), filtered to the relevant status set for each command.

These flags make ko's CLI output directly suitable for HTTP API responses,
removing the need for a backend-for-frontend translation layer.

Check what ko query already emits and match that schema. The key fields the
punchlist client needs: id, title, status, priority, created, deps,
hasUnresolvedDep (computed), description (body), plan-questions.
