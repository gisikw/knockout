---
id: ko-698a
status: open
deps: []
created: 2026-02-24T06:11:23Z
type: task
priority: 2
---
# Support granular tool permissions in pipeline config

## Notes

**2026-02-24 08:47:49 UTC:** ko: FAIL at node 'actionable' — Unresolved architectural decision in Open Questions #1: 'Merge vs Override Semantics: Should allowed_tools at node level completely replace the parent lists, or should it merge (union)?' The plan recommends merge semantics but explicitly requests a decision before implementation.

**2026-02-24 13:02:23 UTC:** Decision: override, not merge. Node-level allowed_tools completely replaces parent lists.
