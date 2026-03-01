## Goal
Add a `--hidden` flag to `ko project set` that causes a project to be excluded from `ko project ls` output.

## Context
The registry is stored in `~/.config/knockout/projects.yml` using a nested YAML format. Each project can have `path`, `prefix`, and `default` properties under its tag key. The `--default` flag is the existing model for setting a per-project boolean property via `ko project set`.

Key files:
- `registry.go`: `Registry` struct, `ParseRegistry`, `FormatRegistry`, `SaveRegistry`, `LoadRegistry`. The struct currently has `Default string`, `Projects map[string]string`, `Prefixes map[string]string`. The pattern for booleans is a single `Default string` (tag name), not a map.
- `cmd_project.go`: `cmdProjectSet` (flag parsing and registry mutation), `cmdProjectLs` (listing with filter logic), `projectJSON` struct (for `--json` output).
- `specs/project_registry.feature`: The spec that must be updated first (per INVARIANTS).
- `testdata/project_registry/*.txtar`: Integration tests via testscript.
- `registry_test.go` and `cmd_project_test.go`: Unit tests.

The INVARIANTS require: spec first, then test, then implementation. Every behavior needs both a Gherkin scenario and a testscript file.

## Approach
Add a `Hidden map[string]bool` field to `Registry`. Extend `ParseRegistry` to read `hidden: true` and `FormatRegistry` to emit it. Add `--hidden` parsing in `cmdProjectSet` and filtering in `cmdProjectLs`. Update the spec and tests accordingly.

## Tasks

1. **[specs/project_registry.feature]** — Add two scenarios: one for setting a project as hidden (`ko project set #tag --hidden` stores `hidden: true`), and one for listing verifying hidden projects are excluded from `ko project ls` output.

2. **[registry.go:Registry]** — Add `Hidden map[string]bool` to the `Registry` struct. Initialize it in `LoadRegistry`'s empty-registry path (alongside the existing `Prefixes` map init).

3. **[registry.go:ParseRegistry]** — In the 4-space indented block handler (new-format nested project properties), add a `case "hidden":` that sets `r.Hidden[currentProject] = (val == "true")`. Initialize `r.Hidden = map[string]bool{}` at the top of `ParseRegistry`.

4. **[registry.go:FormatRegistry]** — In the loop that writes each project's properties, after writing `prefix` and `default`, write `    hidden: true\n` if `r.Hidden[k]` is true.

5. **[cmd_project.go:cmdProjectSet]** — Add `--hidden` to the flag-parsing loop (same pattern as `--default`). After the eviction block, set `reg.Hidden[tag] = true` when the flag is present. Also ensure the eviction block clears `reg.Hidden[existingTag]` when evicting old entries.

6. **[cmd_project.go:cmdProjectLs]** — In both the plain-text and JSON output paths, skip any project where `reg.Hidden[k]` is true. For JSON output, also add `IsHidden bool` to `projectJSON` (though hidden projects won't appear in the list; alternatively, keep `projectJSON` unchanged since hidden projects are simply absent).

7. **[registry_test.go]** — Add `TestParseRegistryHidden` verifying that `hidden: true` is parsed into `reg.Hidden`, and extend `TestFormatRegistryRoundTrip` to include a hidden project and verify round-trip fidelity.

8. **[cmd_project_test.go]** — Add `TestCmdProjectSetHidden` verifying the flag is stored, and `TestCmdProjectLsExcludesHidden` verifying that hidden projects do not appear in `cmdProjectLs` output (capture stdout, check absence).

9. **[testdata/project_registry/hidden_project.txtar]** — New testscript covering: register a project with `--hidden`, verify `ko project ls` does not show it, verify the project is still usable (e.g. routing still works if applicable).

   Verify all: `go test ./...` passes.

## Open Questions

1. **`--all` flag**: Should `ko project ls --all` reveal hidden projects? The ticket doesn't mention it. If hidden projects can never be listed, users lose discoverability of what they've hidden. A `--all` flag is the obvious escape hatch. Flagging for product input — proceeding without it unless asked (hidden means hidden).

2. **Un-hiding**: Is `ko project set #tag` (without `--hidden`) idempotent on the hidden flag, or does it clear it? The `--default` flag is additive-only (you can't un-default via `ko project set`). Likely the same model: `--hidden` sets the flag; clearing it would require a future `--no-hidden` flag or direct file edit. Proceeding with additive-only for consistency.
