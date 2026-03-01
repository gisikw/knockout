## Goal
Replace the flat `projects:`/`prefixes:` sections in `projects.yml` with a single nested structure where each project's `path`, `prefix`, and `default` are grouped under the project tag.

## Context

Current `~/.config/knockout/projects.yml` format:
```yaml
default: exo
projects:
  exo: /home/dev/Projects/exocortex
  knockout: /home/dev/Projects/knockout
prefixes:
  exo: exo
  knockout: ko
```

Target format (all info under the project name, no separate sections):
```yaml
projects:
  exo:
    path: /home/dev/Projects/exocortex
    prefix: exo
    default: true
  knockout:
    path: /home/dev/Projects/knockout
    prefix: ko
```

The in-memory `Registry` struct (Default string, Projects map[string]string, Prefixes map[string]string) is unchanged — only serialization changes.

Key files:
- `registry.go` — `ParseRegistry`, `FormatRegistry`, `LoadRegistry`, `backfillPrefixes`
- `registry_test.go` — unit tests for `ParseRegistry`, `FormatRegistry`
- `cmd_list_test.go` — integration tests that write raw YAML strings to temp files; uses both flat-project and prefixes inline YAML
- `cmd_project_test.go` — tests using `SaveRegistry`/`LoadRegistry` via Go structs (not raw YAML), mostly unaffected

All callers of `reg.Projects[k]`, `reg.Prefixes[k]`, `reg.Default` are in `cmd_project.go`, `cmd_create.go`, `cmd_list.go`, `cmd_serve_sse.go`, and `ticket.go`. They access the struct fields and require no changes — the struct is unchanged.

The parser in `ParseRegistry` is fully manual (no YAML library — zero external deps invariant). It works line-by-line, tracking indentation level and current section.

## Approach

Update `ParseRegistry` to recognize both the old flat format (backward compat) and the new nested format, differentiating based on whether a 2-space-indented `projects:` entry has a value (old: `tag: /path`) or is bare (new: `tag:`). Update `FormatRegistry` to emit the new nested format. Add an `isOldFormat` helper to detect legacy files, and call it in `LoadRegistry` to trigger auto-migration (save in new format on first read of an old-format file). Update all test inline YAML strings to new format.

## Tasks

1. **[registry.go:ParseRegistry]** — Rewrite parser to handle new nested format. When in the `projects:` section and a 2-space-indented line has no value (bare `tag:`), set it as `currentProject`. When `currentProject != ""` and a 4-space-indented line is seen, dispatch on key: `path` → `r.Projects[currentProject]`, `prefix` → `r.Prefixes[currentProject]`, `default: true` → `r.Default = currentProject`. Preserve handling of old-format entries: 2-indent `tag: path` under `projects:` still works, top-level `default: tag` still works, `prefixes:` section still works. A project-name line resets `currentProject`; any top-level or section-header line also resets it.
   Verify: existing `TestParseRegistry`, `TestParseRegistryNoDefault`, `TestParseRegistryBackwardCompatible`, `TestFormatRegistryRoundTrip` all pass with no changes.

2. **[registry.go:FormatRegistry]** — Rewrite to emit the new nested format. Iterate sorted project tags. For each tag, write `  tag:\n`. Write `    path: <path>\n`. If `r.Prefixes[tag]` exists, write `    prefix: <prefix>\n`. If `r.Default == tag`, write `    default: true\n`. Remove the separate `prefixes:` section block and the top-level `default:` key.
   Verify: `TestFormatRegistryRoundTrip` passes; output contains no `prefixes:` section and no top-level `default:`.

3. **[registry.go:isOldFormat + LoadRegistry]** — Add a pure helper `isOldFormat(content string) bool` that returns true if the content contains a top-level `prefixes:` section or a top-level `default:` key (indicating the old format). In `LoadRegistry`, after the `backfillPrefixes` call, also call `isOldFormat(string(data))` and trigger `SaveRegistry` if it returns true (even if backfill made no changes). This auto-migrates the user's existing `~/.config/knockout/projects.yml` on next invocation.
   Verify: a test that loads an old-format file via `LoadRegistry` results in the file being rewritten in new format.

4. **[registry_test.go]** — Update all test inline YAML strings to use new format:
   - `TestParseRegistry`: change input to new nested format, assertions unchanged.
   - `TestParseRegistryNoDefault`: change input to new nested format.
   - `TestFormatRegistryRoundTrip`: the test constructs a `Registry` struct and calls `FormatRegistry` → `ParseRegistry` round-trip; should still pass but verify the output no longer contains `prefixes:`.
   - `TestParseRegistryBackwardCompatible`: rename to `TestParseRegistryOldFormat` (or keep name), ensure it still passes (backward compat must be preserved).
   - Add `TestParseRegistryNewFormat`: explicit test of new nested format with `path`, `prefix`, and `default: true` fields.
   Verify: `go test ./... -run TestParseRegistry` passes.

5. **[cmd_list_test.go]** — Update all inline `regContent` YAML strings from old flat format to new nested format. There are approximately 8 occurrences using `projects:\n  tag: path\n` style (including the one with `prefixes:\n  fn: fn\n`). Convert each to use `  tag:\n    path: ...\n    prefix: ...\n` structure.
   Verify: `go test ./... -run TestResolveProjectTicketsDir` and `go test ./... -run TestCmdTriage` pass.

6. **[Full test suite]** — Run `go test ./...` and confirm all tests pass. The auto-migration logic from task 3 means existing tests that still use old-format YAML will cause `LoadRegistry` to attempt a re-save, but since those tests use temp dirs this is harmless.

## Open Questions

None. All decisions resolved:

- **Auto-migration**: Confirmed — silent auto-migration on first read (task 3 is in scope).
- **Indentation for properties**: 4 spaces (2+2) under a project tag. Consistent with nested YAML style.
- **Empty prefix**: If `reg.Prefixes[tag]` is empty string, skip writing the `prefix:` line.
- **Projects with no default**: If `reg.Default == ""`, no project gets `default: true`.
