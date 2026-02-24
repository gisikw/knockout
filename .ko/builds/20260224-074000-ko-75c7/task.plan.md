I've completed the plan for deprecating the link/unlink commands. Here's what I found:

## Investigation Summary

Links are a symmetric, non-directional association feature between tickets. The implementation touches:

- **Commands**: `ko link` and `ko unlink` (cmd_link.go)
- **Data model**: `Links` field in Ticket struct
- **Display**: Shown in `ko show` and `ko query` output
- **Tests**: Dedicated test suite and spec file
- **Documentation**: README and help text

## Key Finding

Links appear to be completely isolated — they're only used for display and symmetric association tracking. Unlike `deps` (which affect the ready queue and blocking logic), links have no impact on workflow, build pipeline, or any business logic.

## Plan Overview

The plan removes all link functionality in 15 ordered tasks:
1. Remove command routing from main.go
2. Delete cmd_link.go entirely
3. Remove Links field from Ticket struct and all serialization/parsing
4. Remove display in show/query commands
5. Delete tests and spec
6. Clean up documentation

The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-75c7.artifacts/plan.md` and is ready for your review.
