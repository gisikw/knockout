---
id: ko-05bd
status: open
deps: []
created: 2026-02-25T23:48:18Z
type: task
priority: 2
---
# Ensure all output commands support --json flag

Every command that produces human-readable output should also support --json for machine consumption. Audit and add --json where missing. Known commands that need it: show, ls, ready, blocked (if still present), triage (show mode), agent status. Commands that already have it: ls, ready, show. Commands that are write-only (add, close, start, etc) probably just need to output the ticket ID which they already do.
