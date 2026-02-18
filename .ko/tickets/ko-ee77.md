---
id: ko-ee77
status: closed
deps: []
links: []
created: 2026-02-18T19:06:16Z
type: task
priority: 2
---
# Persist prefix to .ko/prefix file — read on create, stop relying on ticket scan

detectPrefix in cmd_create.go scans .tickets/ for existing ticket files to infer the prefix. If the backlog is cleared (all tickets closed/archived/deleted), the next ko create falls through to DerivePrefix(dirName) which may produce a different prefix than what was in use, silently breaking cross-project references. Fix: write a .ko/prefix file when the prefix is first established (either derived or via ko init). On subsequent creates, read .ko/prefix first, only fall back to scan/derive if it doesn't exist. backfillPrefixes in registry.go has the same scan-based detection issue — it should also check .ko/prefix.
