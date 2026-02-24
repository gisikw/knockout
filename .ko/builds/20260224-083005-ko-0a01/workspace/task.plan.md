Perfect! I've written the implementation plan to `/home/dev/Projects/knockout/.ko/tickets/ko-0a01.artifacts/plan.md`.

The plan outlines a straightforward rename operation:
- Add `note` as the new primary command alongside the existing `add-note` alias
- Update all user-facing strings (error messages, help text, README)
- Update test files to use the new name while adding a backward compatibility test
- 7 concrete tasks with verification steps for each

The approach leverages Go's switch statement to make aliasing trivial - both `note` and `add-note` will call the same `cmdAddNote()` function, ensuring perfect backward compatibility during the transition period.
