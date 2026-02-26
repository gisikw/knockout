---
id: ko-88dd
status: blocked
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

**2026-02-26 21:17:23 UTC:** ko: FAIL at node 'review' — Implementation violates two critical invariants: (1) INVARIANTS.md lines 37-41 'Zero external runtime dependencies' - isWorkingTreeClean() shells out to git at runtime, making ko depend on external git binary; (2) INVARIANTS.md lines 14-31 'Every behavior has a spec' - no Gherkin spec added to specs/pipeline.feature. The implementation is otherwise correct and complete, but these invariant violations are blockers that require architectural decisions (use go-git library vs native implementation vs degrade gracefully; spec writing conventions).
