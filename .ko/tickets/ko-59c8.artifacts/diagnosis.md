# Diagnosis: ko project set allows duplicate path registrations

## Symptoms

When `ko project set #newtag` is run in a project directory that is already registered in the global registry under a different tag (e.g. `#oldtag`), both entries persist. The registry ends up with two different tags pointing to the same project root path:

```yaml
projects:
  oldtag: /path/to/project
  newtag: /path/to/project
```

This causes the project to "answer to multiple project tags" — both tag names are valid in lookups, `ko project ls` shows two entries for the same physical project, and cross-project dependency resolution may find the same ticket twice under different tags.

## Root Cause

In `cmdProjectSet` (`cmd_project.go:110`), adding a project to the registry is a pure tag→path map assignment:

```go
reg.Projects[tag] = root
```

This correctly handles:
- **First registration**: inserts the new tag→path pair.
- **Same-tag update** (upsert): overwrites the path for an existing tag key.

What it does NOT handle:
- **New-tag re-registration**: if the same `root` path already has a different `tag` key in `reg.Projects`, the old entry is not removed. Both the old and new tags survive in the map.

There is no scan for existing entries that map to the same `root` before inserting.

The same blind-spot exists for `reg.Prefixes`: the old tag's prefix entry is also left behind.

## Affected Code

| File | Location | Description |
|------|----------|-------------|
| `cmd_project.go` | Lines 97–129, specifically line 110 | `cmdProjectSet` — adds tag→path mapping without evicting stale same-path entries |
| `registry.go` | `Registry` struct, `LoadRegistry`, `SaveRegistry`, `FormatRegistry` | Registry data structures; these are correct as written — the bug is in the write path above |

## Recommended Fix

Before inserting the new `tag → root` mapping, scan `reg.Projects` for any existing key whose value equals `root`. For each such stale key:

1. Delete it from `reg.Projects`.
2. Delete the corresponding entry (if any) from `reg.Prefixes`.
3. If that stale key was `reg.Default`, clear or update `reg.Default` (most conservative: clear it so the user must explicitly re-declare a default with the new tag; alternatively transfer the default to the new tag automatically).

In pseudocode (to be inserted in `cmdProjectSet` just before `reg.Projects[tag] = root`):

```
for existingTag, existingPath := range reg.Projects {
    if existingPath == root && existingTag != tag {
        delete(reg.Projects, existingTag)
        delete(reg.Prefixes, existingTag)
        if reg.Default == existingTag {
            reg.Default = ""  // or: reg.Default = tag
        }
    }
}
```

A corresponding test should cover the scenario:

1. Register the project under `#foo`.
2. Re-register the same directory under `#bar`.
3. Assert `reg.Projects` has exactly one entry (keyed `bar`), not two.
4. Assert `foo` is no longer present.

## Risk Assessment

**Low risk change.** The logic is simple (scan and delete by value), confined to `cmdProjectSet`, and only affects the in-memory `Registry` before `SaveRegistry` is called.

Potential edge cases to be aware of:

- **Default tag clearing**: if the evicted tag was the default, silently clearing `reg.Default` could surprise users who rely on it. Transferring the default to the new tag is more ergonomic but slightly more opinionated. Either behaviour should be documented in the output message.
- **Concurrent writes**: the registry is a flat YAML file with no locking; this is a pre-existing limitation not introduced by the fix.
- **Symlinks / different representations of the same path**: `os.Getwd()` returns the absolute path, so as long as both calls happen from the same canonical path this is fine. If symlinks are involved, `filepath.EvalSymlinks` could be used for a stricter check, but that would be over-engineering for the current use-case.
- **Test coverage**: `TestCmdProjectSetUpsert` in `cmd_project_test.go` and `register_overwrite.txtar` only test same-tag updates; a new test is needed to cover the same-path/different-tag scenario.

---

## Summary

The bug is in `cmdProjectSet` (`cmd_project.go:110`). When a new tag is registered for an already-registered project path, the old tag→path entry is never evicted from the registry map. The fix is a short pre-insertion loop that removes any existing map entries whose value matches the new project root, along with their corresponding `Prefixes` entries. The change is self-contained, low-risk, and requires one new unit test.
