Diagnosis written to `.ko/tickets/ko-59c8.artifacts/diagnosis.md`. Here's the summary:

**Root cause**: `cmdProjectSet` (`cmd_project.go:110`) does a simple map assignment `reg.Projects[tag] = root`. This correctly overwrites an existing entry when the *same tag* is reused, but does nothing when the same *path* already exists under a *different tag*. Both old and new tags persist in the registry, so the project ends up answering to multiple tags.

**Fix**: Before inserting the new tag→path pair, scan `reg.Projects` for any existing key whose value equals `root` (and whose key is not the new tag) and delete those stale entries — plus their corresponding `reg.Prefixes` entries and any `reg.Default` reference.

**Risk**: Very low. The change is a short loop confined to `cmdProjectSet`, only affects the in-memory `Registry` before it's saved, and the only meaningful edge case is deciding what to do with `reg.Default` if the evicted tag was the default (either clear it or transfer it to the new tag).
