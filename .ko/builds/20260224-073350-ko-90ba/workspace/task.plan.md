I've written the plan to `/home/dev/Projects/knockout/.ko/tickets/ko-90ba.artifacts/plan.md`.

The plan outlines a straightforward renaming task:
- Swap the dispatch order in main.go to make `add` primary and `create` the hidden alias
- Update help text to reflect this change
- Update README documentation

The key insight is that both commands will continue calling `cmdCreate` — there's no need to rename internal functions or update tests since both command names will work identically. This preserves backward compatibility while making `ko add` the canonical interface.
