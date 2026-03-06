---
id: ko-12a4
status: open
deps: []
created: 2026-03-05T12:41:17Z
type: task
priority: 2
---
# When a ticket is created in a project that has agent auto start, the agent start message comes through stdout too. I suspect we should make it output to stderr, since it ruins the composability by cluttering stdout

## Notes

**2026-03-06 16:56:25 UTC:** Question: Should we also fix `ko agent stop` to write to stderr for consistency?
Answer: Include in this ticket (Recommended)
Fix both agent start and agent stop in one change for consistency

**2026-03-06 16:56:25 UTC:** Question: Is routing the agent started message to stderr acceptable for direct invocations?
Answer: Yes, standard CLI behavior (Recommended)
Status messages on stderr, data on stdout is the standard CLI pattern
