---
id: ko-a8f7
status: closed
deps: []
created: 2026-04-15T05:15:08Z
type: task
priority: 2
---
# Fix UNIQUE constraint failed on projects.prefix in shadow writes

## Notes

**2026-04-15 05:15:17 UTC:** Tests create tickets with test- prefix in temp dirs. Shadow writes go to global DB (~/.local/state/knockout/knockout.db) which already has projects with conflicting prefixes. The ensureProject function derives prefix from ticketsDir but doesn't handle the case where multiple temp dirs produce the same prefix. Options: (1) namespace test prefixes with random suffix, (2) make prefix non-unique and key on tickets_dir alone, (3) skip shadow writes when ticketsDir is under TMPDIR.

**2026-04-15 05:31:59 UTC:** Testing migration

**2026-04-15 05:32:07 UTC:** Testing migration again

**2026-04-15 05:33:09 UTC:** Migration fix applied

**2026-04-15 05:33:15 UTC:** test
