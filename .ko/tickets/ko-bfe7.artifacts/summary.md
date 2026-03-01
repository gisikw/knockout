## Summary

Restructured `projects.yml` from a flat three-section format (`default:` top-level + `projects:` map + `prefixes:` map) to a single nested format where each project's `path`, `prefix`, and `default` are grouped under the project tag.

### What was done

**`registry.go`**
- `ParseRegistry` extended to handle both old flat format (backward compat) and new nested format. Tracks a `currentProject` variable: bare 2-space keys under `projects:` start a nested block; 4-space-indented `path:`, `prefix:`, `default: true` lines populate that block. Old-format `tag: path` entries still parse correctly.
- `FormatRegistry` rewritten to emit only the new nested format — no separate `prefixes:` section, no top-level `default:` key.
- `isOldFormat(content string) bool` added: detects presence of top-level `prefixes:` or `default:` lines.
- `LoadRegistry` updated to trigger `SaveRegistry` when `isOldFormat` is true, silently auto-migrating old config files on first read.

**Tests**
- `registry_test.go`: Updated `TestParseRegistry` and `TestParseRegistryNoDefault` inputs to new format. Added `TestParseRegistryNewFormat` (explicit nested format coverage). Added assertions to `TestFormatRegistryRoundTrip` verifying no `prefixes:` or top-level `default:` in output. Added `TestLoadRegistryAutoMigrates` verifying the file is rewritten on first read of an old-format config.
- `cmd_list_test.go`: Updated all 8 inline `regContent` YAML strings to new format.
- `testdata/project_registry/register_prefix_detection.txtar`: Updated `stdout` assertions to match new format (`prefix: fn` instead of `prefixes:` + `fort-nix: fn`).

### Notable decisions

- The in-memory `Registry` struct is unchanged — only serialization/deserialization changed, so all callers (`cmd_project.go`, `cmd_list.go`, `ticket.go`, etc.) required no modifications.
- `isOldFormat` checks for unindented `prefixes:` or `default:` lines, so new-format files with nested `default: true` properties (which are 4-space-indented) are never misidentified as old format.
- The `backfillPrefixes` boolean OR with `isOldFormat` ensures auto-migration fires even when no new prefixes are detected.

### Test results

All tests pass: `go test ./...` — `ok git.gisi.network/infra/knockout 10.446s`
