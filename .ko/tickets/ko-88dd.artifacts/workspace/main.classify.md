Looking at this ticket:

**What's being asked:** Add support for a `require_clean_tree: true` configuration option in `pipeline.yml` that prevents pipeline execution when there are uncommitted changes in the working tree.

**Key indicators:**
- Clear feature request: "Add support for..."
- Specific desired behavior documented in notes (blocks when dirty, not when clean)
- Previous implementation attempt failed code review, but **architectural blockers have been resolved** (Feb 28 note grants exception for git shell-out dependency)
- Expected output is committed code implementing this feature

**Why not the other categories:**
- Not research: The question of "what should this do?" was already answered and decided (Feb 26)
- Not bug: Not reporting unexpected behavior, but rather requesting new functionality
- Not blocked: The architectural decisions from the Feb 28 note indicate the way forward is clear

```json
{"disposition": "route", "workflow": "task"}
```
