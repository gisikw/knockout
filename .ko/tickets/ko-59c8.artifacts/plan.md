## Goal
Prevent a project from accumulating multiple registry tags by evicting any stale same-path entry before inserting a new one in `cmdProjectSet`.

## Context
`cmd_project.go:110` does a plain map assignment `reg.Projects[tag] = root`. This is correct for first-registration and same-tag upsert, but does nothing when the same `root` path is already registered under a *different* tag — both entries survive in `reg.Projects` and `reg.Prefixes`.

The `Registry` struct has three relevant fields: `Projects map[string]string`, `Prefixes map[string]string`, and `Default string`. All three must be cleaned up when evicting a stale tag.

Existing test `TestCmdProjectSetUpsert` covers same-tag updates only; no test covers the same-path/different-tag scenario.

## Approach
Just before `reg.Projects[tag] = root` in `cmdProjectSet`, iterate over `reg.Projects` and delete any entry whose value equals `root` but whose key is not `tag`. Delete the corresponding `reg.Prefixes` entry too. If the evicted tag was `reg.Default`, transfer the default to the incoming `tag` (more ergonomic than silently clearing it, since the user is explicitly re-tagging the same project).

## Tasks
1. [`cmd_project.go:110`] — Insert a pre-eviction loop immediately before `reg.Projects[tag] = root`. For each `existingTag, existingPath` in `reg.Projects`: if `existingPath == root && existingTag != tag`, delete `reg.Projects[existingTag]`, delete `reg.Prefixes[existingTag]`, and if `reg.Default == existingTag` set `reg.Default = tag`.
   Verify: `go test ./...` passes with no regressions.

2. [`cmd_project_test.go`] — Add `TestCmdProjectSetRetagEvictsOldTag`: register the same `projectDir` under `#foo`, then re-register under `#bar`. Assert `len(reg.Projects) == 1`, that key `"bar"` exists with the correct path, and that key `"foo"` is absent. Add a second sub-case: if `#foo` was the default, assert `reg.Default` is now `"bar"` after re-registering as `#bar`.
   Verify: new test passes, existing tests unbroken.

## Open Questions
- **Default transfer vs. clear**: plan assumes the default is *transferred* to the new tag when the evicted tag was the default. The alternative (clear `reg.Default` to `""`) is safer but more surprising. Either is defensible; transfer is chosen here as the more ergonomic option since the user is explicitly re-tagging the same project. Implementer may override if the product preference differs.
