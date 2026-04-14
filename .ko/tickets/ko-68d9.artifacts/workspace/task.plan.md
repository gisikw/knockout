knockout — just --list for recipes
**Plan written to `.ko/tickets/ko-68d9.artifacts/plan.md`.**

Here's the summary:

**Goal**: Merge `projects.yml` into `config.yaml` as a `projects:` section.

**Key findings**:
- Both parsers (`ParseRegistry`, `ParseGlobalConfig`) already tolerate each other's sections — the only real problem is `SaveRegistry` overwrites the whole file with just the projects block.
- `RegistryPath()` is the single chokepoint: 9 call sites, all just need the return value to be `config.yaml`.
- Tests in `cmd_project_test.go` hardcode `"projects.yml"` in 13 places.

**Approach** (6 tasks):
1. `RegistryPath()` → delegate to `GlobalConfigPath()`
2. New `mergeProjectsSection()` pure function — splices the `projects:` block into the config file while preserving other settings
3. `SaveRegistry()` → use merge-save instead of overwrite
4. `LoadRegistry()` → transparent migration: if `config.yaml` has no projects and old `projects.yml` exists, read it and save to `config.yaml`
5. New tests for merge and migration
6. Update 13 `regPath` references in `cmd_project_test.go`

**Open questions**: (1) Is leaving `projects.yml` on disk acceptable (not auto-deleting it)? (2) Should we test the edge case of settings placed *after* the projects block?
