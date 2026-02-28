Plan updated. Here's a summary of what changed from the prior plan:

**Resolved issues:**
- The git dependency concern is addressed — the ticket author granted an explicit exception on 2026-02-28, noting that `require_clean_tree` runs in the same pipeline context as `on_succeed`/`on_fail` hooks which already assume git is available.

**New tasks added (the root cause of the previous review failure):**
- **Task 9** — Add Gherkin scenarios to `specs/pipeline.feature` in the Eligibility section (two scenarios: dirty tree blocked, `.ko/`-only changes allowed)
- **Task 10** — Add `testdata/pipeline/build_require_clean_tree.txtar` integration test that initializes a real git repo, stages changes, and verifies the feature end-to-end via `ko agent build`

**Existing tasks 1–8** are unchanged from the prior implementation (which was functionally correct — it only failed review due to the spec/invariant gaps). Task 11 (example files) was task 8 previously.
