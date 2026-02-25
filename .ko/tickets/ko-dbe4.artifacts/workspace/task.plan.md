Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-dbe4.artifacts/plan.md`.

The investigation shows this is a clean removal:
- The `add-note` alias exists only in the main switch statement (lines 52-53)
- There's one backward compatibility test that should be removed
- Documentation already shows only `note`, so no updates needed
- The change is straightforward with no architectural decisions required
