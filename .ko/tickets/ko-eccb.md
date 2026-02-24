---
id: ko-eccb
status: open
deps: [ko-8a7c, ko-f3e2, ko-3b00]
links: []
created: 2026-02-24T03:04:21Z
type: task
priority: 2
---
# Review node: check plan compliance, write after-action summary

Write the `review.md` prompt for the task workflow's review node.

The review node reads the plan artifact and compares the git diff against it.
Beyond the existing review criteria (correctness, completeness, safety, scope,
tests, invariants), it also checks:

- Plan compliance: were all planned tasks completed?
- Deviation detection: if the implementation diverged from the plan, is
  the deviation explained and justified?
- Risk resolution: did the risks/assumptions from the plan materialize?
  Were they handled?

After review, writes a brief after-action summary to the artifact dir.
This gets picked up by the on_succeed commit message so the reasoning
survives in git history after artifacts are cleaned up.

Depends on: ko-8a7c (artifact directory), ko-f3e2 (plan node)
