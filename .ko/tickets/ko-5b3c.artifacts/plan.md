## Goal
Add a `--hidden` flag to `ko project set` that excludes a project from `ko project ls`, with `ko project ls --all` revealing hidden projects alongside normal ones.

## Context
The registry is stored in `~/.config/knockout/projects.yml` using a nested YAML format. Each project can have `path`, `prefix`, and `default` properties under its tag key. The `--default` flag is the existing model for setting a per-project boolean property via `ko project set`.

Key files:
- `registry.go`: `Registry` struct, `ParseRegistry`, `FormatRegistry`, `SaveRegistry`, `LoadRegistry`. The struct currently has `Default string`, `Projects map[string]string`, `Prefixes map[string]string`. Boolean flags per-project use a map (see `Prefixes`), not a single string.
- `cmd_project.go`: `cmdProjectSet` (flag parsing, hand-rolled loop over args), `cmdProjectLs` (uses `flag.NewFlagSet`, currently only has `--json`), `projectJSON` struct (for `--json` output).
- `specs/project_registry.feature`: The spec that must be updated first (per INVARIANTS).
- `testdata/project_registry/*.txtar`: Integration tests via testscript.
- `registry_test.go` and `cmd_project_test.go`: Unit tests.

The INVARIANTS require: spec first, then test, then implementation. Every behavior needs both a Gherkin scenario and a testscript file.

Open question resolved: hidden projects are revealed via `ko project ls --all` (ticket notes, 2026-03-01).

## Approach
Add `Hidden map[string]bool` to `Registry`. Extend `ParseRegistry`/`FormatRegistry` to handle `hidden: true` as a 4-space-indented project property. Add `--hidden` to `cmdProjectSet`'s arg loop and `--all` to `cmdProjectLs`'s `flag.NewFlagSet`. Filter hidden projects from `cmdProjectLs` output unless `--all` is passed.

## Tasks

1. **[specs/project_registry.feature]** — Add three scenarios:
   - `ko project set #tag --hidden` stores `hidden: true` in the registry YAML.
   - `ko project ls` excludes hidden projects from output.
   - `ko project ls --all` includes hidden projects in output (alongside normal ones).
   Verify: spec file is valid Gherkin (review manually).

2. **[registry.go:Registry]** — Add `Hidden map[string]bool` field. Initialize it alongside `Prefixes` in `LoadRegistry`'s empty-registry return (`&Registry{Projects: map[string]string{}, Prefixes: map[string]string{}, Hidden: map[string]bool{}}`).
   Verify: `go build ./...` succeeds.

3. **[registry.go:ParseRegistry]** — Initialize `r.Hidden = map[string]bool{}` at the top of `ParseRegistry` (alongside `Projects` and `Prefixes`). In the 4-space-indented property switch, add `case "hidden": if val == "true" { r.Hidden[currentProject] = true }`.
   Verify: existing parse tests still pass.

4. **[registry.go:FormatRegistry]** — In the per-project loop, after writing `default: true` (if applicable), write `    hidden: true\n` if `r.Hidden[k]` is true.
   Verify: round-trip test passes (format then parse gives same struct).

5. **[cmd_project.go:cmdProjectSet]** — Add `--hidden` to the hand-rolled flag-parsing loop (same `else if arg == "--hidden"` pattern as `--default`). In the eviction block, also `delete(reg.Hidden, existingTag)`. After setting `reg.Projects[tag] = root`, set `reg.Hidden[tag] = true` when the flag was present. Update the success message to mention "hidden" when appropriate.
   Verify: `go build ./...` succeeds.

6. **[cmd_project.go:cmdProjectLs]** — Add `allProjects := fs.Bool("all", false, "include hidden projects")` to the `flag.NewFlagSet` setup. In both the text and JSON output paths, skip entries where `reg.Hidden[k]` is true unless `*allProjects` is true. Add `IsHidden bool` to `projectJSON` so JSON output marks hidden projects when `--all` is used (instead of silently omitting the field). Update the usage string in `cmdProject` to note `--all`.
   Verify: `go build ./...` succeeds.

7. **[registry_test.go]** — Add `TestParseRegistryHidden`: input YAML with `hidden: true` on one project, assert `reg.Hidden["tag"]` is true and other tags are not present in `Hidden`. Extend `TestFormatRegistryRoundTrip` to include a hidden project and verify the round-trip preserves it.
   Verify: `go test ./... -run TestParseRegistryHidden` and round-trip test pass.

8. **[cmd_project_test.go]** — Add:
   - `TestCmdProjectSetHidden`: call `cmdProjectSet` with `--hidden`, load the registry, assert `reg.Hidden["tag"]` is true.
   - `TestCmdProjectLsExcludesHidden`: register a hidden project, capture stdout from `cmdProjectLs`, assert tag is absent.
   - `TestCmdProjectLsAllShowsHidden`: register a hidden project, capture stdout from `cmdProjectLs([]string{"--all"})`, assert tag is present.
   Verify: `go test ./... -run TestCmdProject` passes.

9. **[testdata/project_registry/hidden_project.txtar]** — New testscript:
   - Register a project with `ko project set #secret --hidden`.
   - Run `ko project ls` and assert `secret` does not appear.
   - Run `ko project ls --all` and assert `secret` appears.
   - Optionally verify the YAML file contains `hidden: true`.
   Verify: `go test ./...` passes (testscript tests run as part of the test suite).

## Open Questions

None. The `--all` flag question was answered in the ticket notes (2026-03-01). Un-hiding is additive-only (consistent with `--default`): `--hidden` sets the flag; clearing it requires direct file edit or a future `--no-hidden` flag.
