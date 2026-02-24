---
id: ko-8a7c
status: closed
deps: []
links: []
created: 2026-02-24T03:04:21Z
type: task
priority: 2
---
# Artifact directory per ticket (ko-XXXX.artifacts/)

Pipeline nodes need a place to persist intermediate work products — plans,
investigation notes, triage context — that downstream nodes can optionally
pull rather than re-deriving.

`.ko/tickets/ko-XXXX.artifacts/` is created on first build, available to
all nodes via an env var, and cleaned up on ticket close.

This is separate from the JSONL build log (ko-6d2a) which tracks mechanical
build history. Artifacts are "what the agent learned and decided." JSONL is
"what happened."

Nodes write artifacts with predictable names (e.g. `plan.md`, `triage-context.md`).
Prompts can reference prior artifacts to avoid redundant codebase crawls.

On close: delete the directory. The reasoning lives in git history (commit
messages), not in artifact cruft that grows forever.
