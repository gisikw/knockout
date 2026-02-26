---
id: ko-88dd
status: open
deps: []
created: 2026-02-26T20:30:31Z
type: task
priority: 2
---
# Add support for require_clean_tree: true in pipeline.yml, which would prevent an agent from starting if the working tree is clean (ignoring ./.ko)

## Notes

**2026-02-26 21:05:13 UTC:** Question: Should require_clean_tree: true prevent builds when there ARE uncommitted changes, or when there are NO uncommitted changes?
Answer: Prevent if dirty (Recommended)
require_clean_tree: true blocks builds when there ARE uncommitted changes (standard interpretation)
