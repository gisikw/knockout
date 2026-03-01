# After-Action Summary: ko-59c8

## What was done

Added a pre-eviction loop to `cmdProjectSet` in `cmd_project.go` (just before the `reg.Projects[tag] = root` assignment at line ~110). The loop iterates over all existing registry entries and, for any entry whose path matches the incoming `root` but whose key differs from the new `tag`, deletes both the `reg.Projects` and `reg.Prefixes` entries for that stale tag, and transfers `reg.Default` to the new tag if the evicted tag was the default.

## Tests

`TestCmdProjectSetRetagEvictsOldTag` was added to `cmd_project_test.go` with two sub-tests:
- `retag removes old entry`: registers under `#foo`, re-registers under `#bar`, asserts exactly one entry survives keyed `"bar"` with `"foo"` absent.
- `retag transfers default`: registers under `#foo --default`, re-registers under `#bar`, asserts `reg.Default == "bar"`.

All tests pass.

## Notable decisions

The plan's open question (transfer vs. clear default) was resolved before implementation reached this stage: transfer the default to the new tag, since the user is explicitly re-tagging the same project and losing the default would be surprising.

## Fix applied during verification

The implementation was missing Gherkin specs for the two new behaviors, violating the INVARIANTS.md contract ("Every behavior has a spec"). Two scenarios were added to `specs/project_registry.feature`:
- "Re-registering under a different tag evicts the old entry"
- "Default transfers to new tag when re-registering"
