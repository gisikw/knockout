---
id: ko-3639
status: open
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Make specs runnable with behave

The specs/ directory has .feature files but they're not wired up to run.
Introduce behave (Python BDD framework) as a test dependency and implement
step definitions so the specs actually execute.

Alternatively, evaluate whether the txtar-based tests already cover the
same ground and the .feature files are redundant. If so, pick one approach
and commit to it.
