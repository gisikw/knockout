All 6 tasks complete. Here's a summary of what changed:

---

**Task 1 & 2 — `registry.go`: `ParseRegistry` and `FormatRegistry`**

- `ParseRegistry` now handles both the old flat format (backward compat) and the new nested format. It tracks a `currentProject` variable: when it encounters a 2-space-indented bare key under `projects:` (e.g. `  exo:` with no value), it sets `currentProject` and then dispatches 4-space-indented `path:`, `prefix:`, and `default: true` lines to the appropriate struct fields. Old-format entries (`  exo: /path`) still work — the `val != ""` check distinguishes them.
- `FormatRegistry` now emits the new nested format: `projects:\n  <tag>:\n    path: ...\n    prefix: ...\n    default: true\n` — no separate `prefixes:` section, no top-level `default:` key.

**Task 3 — `registry.go`: `isOldFormat` + `LoadRegistry`**

- Added `isOldFormat(content string) bool` which returns true if the content has a top-level `prefixes:` or `default:` line.
- `LoadRegistry` now triggers `SaveRegistry` when `isOldFormat` is true (in addition to when `backfillPrefixes` makes changes), silently auto-migrating old-format files on first read.

**Task 4 — `registry_test.go`**

- Updated `TestParseRegistry` and `TestParseRegistryNoDefault` inputs to use the new nested format.
- Added `TestParseRegistryNewFormat`: explicit test of new nested format with `path`, `prefix`, and `default: true` fields.
- Added assertions to `TestFormatRegistryRoundTrip` that the output does not contain `prefixes:` or start with `default:`.
- Kept `TestParseRegistryBackwardCompatible` testing old flat format (backward compat preserved).
- Added `TestLoadRegistryAutoMigrates`: loads an old-format file and verifies the file is rewritten in new format.
- Added `"os"`, `"path/filepath"`, and `"strings"` imports.

**Task 5 — `cmd_list_test.go`**

Updated 8 inline `regContent` YAML strings from old flat format to new nested format, including the cross-project test that had a `prefixes:` section.

**Task 6 — txtar test (`testdata/project_registry/register_prefix_detection.txtar`)**

Updated the `stdout` assertions from `stdout 'prefixes:'` + `stdout 'fort-nix: fn'` to `stdout 'prefix: fn'` to match the new nested format.

**No deviations from the plan.** All tests pass (`go test ./...`).
