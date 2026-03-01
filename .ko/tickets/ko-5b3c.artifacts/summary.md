# Summary

## What was done

Added a `--hidden` flag to `ko project set` and a `--all` flag to `ko project ls`, as specified in the ticket and plan.

**Registry layer (`registry.go`):**
- Added `Hidden map[string]bool` field to `Registry` struct
- `ParseRegistry` initializes `Hidden` and reads `hidden: true` from nested YAML project blocks
- `FormatRegistry` emits `    hidden: true` for marked projects
- `LoadRegistry`'s empty-registry return initializes `Hidden` alongside `Prefixes`

**CLI layer (`cmd_project.go`):**
- `cmdProjectSet`: `--hidden` flag parsed in the hand-rolled arg loop; eviction block clears `reg.Hidden[existingTag]`; sets `reg.Hidden[tag] = true` when flag is present; success message updated
- `cmdProjectLs`: `--all` flag added via `flag.NewFlagSet`; hidden projects skipped in both text and JSON paths unless `--all` is set; `IsHidden bool` added to `projectJSON`; usage strings updated

**Spec (`specs/project_registry.feature`):** Three Gherkin scenarios added:
1. `ko project set #tag --hidden` stores `hidden: true` in registry YAML
2. `ko project ls` excludes hidden projects
3. `ko project ls --all` includes hidden projects

**Tests:**
- `registry_test.go`: `TestParseRegistryHidden` added; `TestFormatRegistryRoundTrip` extended with a hidden project
- `cmd_project_test.go`: `TestCmdProjectSetHidden`, `TestCmdProjectLsExcludesHidden`, `TestCmdProjectLsAllShowsHidden` added
- `testdata/project_registry/hidden_project.txtar`: Integration testscript covering the full CLI flow

All 9 planned tasks were completed. `go test ./...` passes.

## Notable decisions

- **`--all` flag**: Answered by ticket notes before this run — hidden projects are revealed with `ko project ls --all`, not permanently unlisted.
- **Un-hiding is additive-only**: `--hidden` sets the flag; clearing it requires direct file edit or a future `--no-hidden` flag. Consistent with how `--default` works.
- **`IsHidden bool` in `projectJSON`**: Included so JSON consumers can distinguish hidden projects when `--all` is used, rather than silently omitting the field.

## Future reader note

`cmd_project_test.go` is now at 566 lines, crossing the 500-line INVARIANT threshold (it was 464 before this change). The INVARIANTS call for ticketing the split rather than blocking work; the test mirrors its source file (`cmd_project.go`) so a separate test file isn't appropriate without first splitting the source. A ticket to split `cmd_project.go` along behavioral seams would bring both files back into compliance.
