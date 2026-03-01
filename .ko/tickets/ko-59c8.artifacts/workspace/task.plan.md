Plan written to `.ko/tickets/ko-59c8.artifacts/plan.md`. Here's a summary:

**Goal:** Evict stale registry entries before inserting a new tag for an already-registered project path.

**Root cause:** `cmd_project.go:110` does `reg.Projects[tag] = root` with no prior scan for existing entries pointing to the same `root`. Both old and new tags survive.

**Fix (one place):** Add a short loop before that assignment in `cmdProjectSet` — iterate `reg.Projects`, delete any `existingTag → root` pair where `existingTag != tag`, clean up the matching `reg.Prefixes` entry, and transfer `reg.Default` if it was the evicted tag.

**Test:** New `TestCmdProjectSetRetagEvictsOldTag` — register under `#foo`, re-register under `#bar`, assert exactly one entry remains keyed `"bar"` with `"foo"` gone. Second sub-case covers default-transfer.

**One open question:** whether the default should be *transferred* to the new tag (chosen here) or *cleared* when the evicted tag was the default.
