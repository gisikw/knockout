---
id: ko-59c8
status: open
deps: []
created: 2026-03-01T06:14:41Z
type: task
priority: 2
---
# If `ko project set` is run on a project that already has an entry in the registry, that prior entry should be replaced. The project shouldn't answer to multiple project tags

## Notes

**2026-03-01 06:39:36 UTC:** Question: When re-registering a project under a different tag, should the default be transferred to the new tag or cleared?
Answer: Transfer default to new tag (Recommended)
Set reg.Default to the new tag. More ergonomic since the user is explicitly re-tagging the project.
