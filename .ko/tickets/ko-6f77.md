---
id: ko-6f77
status: open
deps: []
created: 2026-02-25T23:17:36Z
type: task
priority: 2
---
# Add --project flag, strip hashtag auto-detection from add

Commands that need --project (no ticket ID to derive context from): add, ls, ready, agent loop, agent init, agent start, agent stop, agent status. Also strip RouteTicket hashtag parsing from title in ko add — use --project exclusively. Implementation: modify resolveProjectTicketsDir() to accept an optional --project flag value.
