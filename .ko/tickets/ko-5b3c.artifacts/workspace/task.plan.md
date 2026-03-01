The plan is written to `.ko/tickets/ko-5b3c.artifacts/plan.md`. Here's a summary:

**Goal**: Add `--hidden` flag to `ko project set` so hidden projects are excluded from `ko project ls`.

**Key findings**:
- `Registry` struct in `registry.go` uses separate maps per attribute (`Projects`, `Prefixes`). A `Hidden map[string]bool` field follows the same pattern.
- `ParseRegistry`/`FormatRegistry` handle the nested YAML format — adding `hidden: true` as a 4-space-indented property is straightforward.
- `cmdProjectSet` already has a flag-parsing loop for `--default` and `--prefix=` — `--hidden` slots in identically.
- `cmdProjectLs` needs a skip condition in both the text and JSON paths.
- Per INVARIANTS: spec (`specs/project_registry.feature`) must be written first, then a testscript in `testdata/project_registry/`, then the implementation.

**Open questions flagged in the plan**:
1. Should `ko project ls --all` reveal hidden projects? (No mention in ticket)
2. Is un-hiding (`ko project set #tag` without `--hidden`) expected to clear the flag? (Proceeding with additive-only, consistent with `--default`)
