---
id: ko-5b3c
status: open
deps: [ko-bfe7]
created: 2026-03-01T06:18:30Z
type: task
priority: 2
---
# Add support for a "hidden" flag in projects (which can be set similarly to the --default flag). If a project is hidden, it should be excluded from ko project ls

## Notes

**2026-03-01 13:40:38 UTC:** Question: Should hidden projects be revealed with a `--all` flag, or remain permanently unlisted?
Answer: Reveal with --all flag
Add `ko project ls --all` to show hidden projects alongside normal ones

**2026-03-01 13:40:41 UTC:** ko: FAIL at node 'build' — unknown workflow 'main'
