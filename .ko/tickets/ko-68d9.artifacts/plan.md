## Goal
Consolidate `projects.yml` into `config.yaml` so the project registry lives in a single config file.

## Context

- `registry.go`: `RegistryPath()` returns `~/.config/knockout/projects.yml`. `LoadRegistry(path)` reads/parses it. `SaveRegistry(path, reg)` overwrites the file with only the projects YAML block. `ParseRegistry` handles the `projects:` section and ignores unknown top-level keys (e.g. `summarizer`).
- `global_config.go`: `GlobalConfigPath()` returns `~/.config/knockout/config.yaml`. `ParseGlobalConfig` handles `summarizer:` and ignores `projects:` and nested indented lines.
- `RegistryPath()` is called in 9 places: `ticket.go`, `cmd_agent.go`, `cmd_create.go`, `cmd_serve_sse.go` (×2), `cmd_import.go` (×2), `cmd_project.go` (×2), `cmd_list.go` (×2).
- Both parsers are already tolerant of extra sections; the only problem is `SaveRegistry` overwrites the whole file with only the projects block.
- Tests in `cmd_project_test.go` explicitly construct `regPath` as `filepath.Join(dir, "knockout", "projects.yml")` — 13 occurrences need updating.

## Approach

Change `RegistryPath()` to delegate to `GlobalConfigPath()`. Update `SaveRegistry` to do a section-aware merge — preserve non-project content in `config.yaml` and replace only the `projects:` block. Add transparent migration in `LoadRegistry`: if parsing `config.yaml` yields an empty registry and the old `projects.yml` exists, load from `projects.yml` and immediately save back to `config.yaml` so the migration persists.

## Tasks

1. **[registry.go:RegistryPath]** — Change the function body to `return GlobalConfigPath()`. No other callers change; they all already call `RegistryPath()`.
   Verify: `grep projects.yml registry.go` returns nothing except comments/docs.

2. **[registry.go:mergeProjectsSection (new function)]** — Add a pure function `mergeProjectsSection(existingConfig, newProjects string) string` that:
   - Scans `existingConfig` line-by-line to find the first top-level `projects:` line (no leading whitespace).
   - Collects lines before it (header) and lines after the projects block ends (footer — next top-level non-empty, non-comment line onward).
   - Returns `header + newProjects + footer`. If no `projects:` line exists, appends `newProjects` after a trailing newline.
   Verify: unit tests in `registry_test.go` (task 4) cover header-only, projects-only, header+projects+footer, and empty-config cases.

3. **[registry.go:SaveRegistry]** — Replace the direct `os.WriteFile(path, []byte(FormatRegistry(r)), 0644)` with:
   - Read existing config file (empty string if absent).
   - Call `mergeProjectsSection(existing, FormatRegistry(r))`.
   - Write the merged result.
   Verify: `go test ./...` passes; `TestFormatRegistryRoundTrip` still passes; manual check that `summarizer:` in config.yaml survives a `ko project set`.

4. **[registry.go:LoadRegistry]** — After `ParseRegistry` returns, if `len(reg.Projects) == 0`, check whether the old `projects.yml` path exists (compute old path: same dir as `GlobalConfigPath()` but filename `projects.yml`). If it exists, load and parse it, then call `SaveRegistry` to migrate. Return the migrated registry.
   Verify: test `TestLoadRegistryMigration` (task 5) passes.

5. **[registry_test.go]** — Add two tests:
   - `TestMergeProjectsSection`: checks that `mergeProjectsSection` correctly splices the projects block in while preserving header/footer settings lines.
   - `TestLoadRegistryMigration`: creates a temp dir with an old `projects.yml` (no `config.yaml`), calls `LoadRegistry(configYamlPath)`, verifies the returned registry has the projects and that `config.yaml` now exists with the projects section (migration persisted).
   Verify: new tests pass.

6. **[cmd_project_test.go]** — Update all 13 `regPath := filepath.Join(dir, "knockout", "projects.yml")` occurrences to `filepath.Join(dir, "knockout", "config.yaml")`. Also update any `SaveRegistry(regPath, reg)` call sites that use the old path — they already use `regPath` so only the path string changes.
   Verify: `go test ./... -run TestCmdProject` passes.

## Open Questions

- **Migration destructiveness**: The plan leaves `projects.yml` in place after migration (non-destructive). Obligation 2 says it "no longer exists OR is no longer read" — "no longer read" is satisfied since after one migration write the code always reads `config.yaml`. No user action required. Acceptable?
- **Footer ordering**: If a user has manually placed non-project settings after the `projects:` block in `config.yaml`, `mergeProjectsSection` will preserve them as footer. In practice only `summarizer:` exists as a setting and it typically precedes projects, but this edge case should be explicitly tested.
