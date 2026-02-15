---
id: ko-54fb
status: open
deps: []
links: []
created: 2026-02-15T04:59:15Z
type: task
priority: 3
---
# Build retry logic for non-signal failures

Currently retries are wired but only for execution errors. Need to verify retry behavior for prompt stages that exit non-zero without an explicit signal (FAIL/BLOCKED/DECOMPOSE). Add tests.
