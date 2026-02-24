---
id: ko-6d2a
status: open
deps: []
links: []
created: 2026-02-24T03:04:21Z
type: task
priority: 2
---
# JSONL build history per ticket (ko-XXXX.jsonl)

Separate from the per-build event log (which lives in `.ko/builds/`), each
ticket gets a persistent JSONL file tracking its mechanical build history
across retries.

`.ko/tickets/ko-XXXX.jsonl` — append-only, one line per event:
- Build started (timestamp, pipeline version)
- Node entered/exited (name, outcome, duration)
- Disposition emitted (type, reason)
- Retry triggered (attempt number, which node failed)
- Build completed (outcome: succeed/fail/blocked/decompose)

This survives across builds so you can see the full history: "picked up at
3:47, triage passed, implement failed, retried, implement passed, verify
failed, blocked." Queryable with `jq`. Readable at a glance with `cat`.

Unlike the artifact directory, this persists after close — it's the audit
trail.
