All 9 tasks complete. Here's a summary of changes:

---

## Summary

### Task 1 — `specs/project_registry.feature`
Added three Gherkin scenarios: setting a project as hidden stores `hidden: true` in YAML; `ko project ls` excludes hidden projects; `ko project ls --all` includes hidden projects.

### Task 2 — `registry.go:Registry`
Added `Hidden map[string]bool` field to the `Registry` struct. Updated `LoadRegistry`'s empty-registry return to initialize `Hidden: map[string]bool{}`.

### Task 3 — `registry.go:ParseRegistry`
Initialized `Hidden: map[string]bool{}` in the `Registry` literal at the top of `ParseRegistry`. Added `case "hidden"` to the 4-space-indented property switch.

### Task 4 — `registry.go:FormatRegistry`
Added `hidden: true\n` serialization after `default: true` in the per-project loop, gated on `r.Hidden[k]`.

### Task 5 — `cmd_project.go:cmdProjectSet`
Added `setHidden bool` variable and `else if arg == "--hidden"` branch to the arg loop. Added `delete(reg.Hidden, existingTag)` in the eviction block. Added `reg.Hidden[tag] = true` when `setHidden` is true. Updated success message and usage strings.

### Task 6 — `cmd_project.go:cmdProjectLs`
Added `IsHidden bool` to `projectJSON`. Added `allProjects := fs.Bool("all", false, ...)` to the FlagSet. Added hidden-filtering in both text and JSON output paths (skip if `reg.Hidden[k] && !*allProjects`). JSON output now includes `IsHidden` field. Updated usage strings.

### Task 7 — `registry_test.go`
Added `TestParseRegistryHidden`. Extended `TestFormatRegistryRoundTrip` to include a hidden project and assert the round-trip preserves it.

### Task 8 — `cmd_project_test.go`
Added `TestCmdProjectSetHidden`, `TestCmdProjectLsExcludesHidden`, and `TestCmdProjectLsAllShowsHidden`. Added `"strings"` to imports.

### Task 9 — `testdata/project_registry/hidden_project.txtar`
New testscript: registers a hidden project + a visible one, asserts `ko project ls` excludes the hidden one, asserts `ko project ls --all` includes both, and verifies `hidden: true` appears in the YAML.

**No deviations from the plan.**
