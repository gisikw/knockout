---
id: ko-bfe7
status: open
deps: []
created: 2026-03-01T06:17:26Z
type: task
priority: 2
---
# Restructure the projects.yml to avoid redundant lists of things. Everything (path, default, prefix) can just be aggregated under a project name. Update everywhere in knockout that has dependencies on this file structure

## Notes

**2026-03-01 06:39:19 UTC:** Question: Should knockout auto-migrate users' existing projects.yml files from the old format to new format on first read, or should users manually migrate their config?
Answer: Silent auto-migration (Recommended)
Implement task 3: auto-detect old format and rewrite config file automatically on first read
