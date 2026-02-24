## Goal
Consolidate `.ko/` top-level configuration files into a single `.ko/config.yaml`.

## Context

Currently `.ko/` has several scattered files and configuration:

**Files to consolidate:**
- `pipeline.yml` — Pipeline configuration (workflows, hooks, model, discretion, etc.)
- `prefix` — Project ticket prefix (read/written by `cmd_create.go:ReadPrefix/WritePrefix`)

**Files to keep as-is (runtime state, not config):**
- `agent.lock` — Runtime flock for `ko agent loop` (cmd_agent.go:84-98)
- `agent.pid` — Runtime PID file for background agent (cmd_agent.go:49-71)
- `agent.log` — Runtime log output (cmd_agent.go:54-57)

**Current code patterns:**
- Pipeline config is loaded via `pipeline.go:FindPipelineConfig` → `LoadPipeline` → `ParsePipeline`
- The `Pipeline` struct (pipeline.go:11-25) contains all pipeline-level config
- Prefix is read/written via dedicated functions in `cmd_create.go:159-178`
- Registry (`~/.config/knockout/projects.yml`) also tracks prefixes (registry.go:14) but `.ko/prefix` takes precedence for local projects
- Tests expect `.ko/prefix` to exist (cmd_init_test.go:25-27)

**Architectural constraints (from INVARIANTS.md):**
- No external YAML dependencies — use minimal hand-rolled parser like `ParsePipeline` does
- Backwards compatibility not required (INVARIANTS.md:169-170 "avoid backwards-compatibility hacks")
- Keep functions pure; I/O separate from logic (INVARIANTS.md:110-120)

**Related tickets:**
- ko-1930 depends on this ticket to decide where agent harness overrides live in project config

## Approach

Create a unified `.ko/config.yaml` that contains both pipeline configuration and project-level settings (prefix, etc.). Extend the existing `Pipeline` struct to include a `Prefix` field, and update the parser to handle both sections. Update all code that reads `pipeline.yml` or `.ko/prefix` to use the new config file. Maintain the same minimal YAML parsing approach used in `pipeline.go:ParsePipeline`.

The config structure will be:
```yaml
# Project settings
prefix: ko

# Pipeline settings (existing fields)
model: claude-sonnet-4-5-20250929
discretion: medium
# ... workflows, on_succeed, etc.
```

## Tasks

1. [pipeline.go:Pipeline] — Add `Prefix string` field to the `Pipeline` struct.
   Verify: `go build` succeeds.

2. [pipeline.go:ParsePipeline] — Extend the YAML parser to recognize `prefix:` as a top-level scalar field and populate `p.Prefix`.
   Verify: unit test confirms parsing works.

3. [pipeline.go:FindPipelineConfig] — Update to look for `config.yaml` first, then fall back to `pipeline.yml` for backwards compatibility during transition. Update function comment.
   Verify: function returns correct path for both old and new config files.

4. [pipeline_test.go] — Add test case for parsing config with `prefix:` field.
   Verify: `go test ./... -run TestParsePipeline` passes.

5. [cmd_create.go:ReadPrefix] — Update to read from pipeline config (`Prefix` field) instead of `.ko/prefix` file. Handle case where config doesn't exist (return empty string).
   Verify: reading prefix from config works.

6. [cmd_create.go:WritePrefix] — Update to write prefix to config file instead of standalone `.ko/prefix`. This requires reading existing config, updating the prefix field, and writing back. Consider whether this is needed or if prefix should only be set during init.
   Verify: prefix persistence works correctly.

7. [cmd_create.go:detectPrefix] — Update to check pipeline config first before scanning ticket files. Update function comment.
   Verify: prefix detection follows new precedence order.

8. [cmd_init.go:cmdInit] — Update to write prefix to `.ko/config.yaml` instead of `.ko/prefix` file. Consolidate with `cmd_agent.go:cmdAgentInit` config writing if both are setting the same file.
   Verify: `ko init` creates config with prefix.

9. [cmd_build_init.go:cmdAgentInit] — Rename generated file from `pipeline.yml` to `config.yaml` in the default template. Update prompts that reference the filename.
   Verify: `ko agent init` creates config.yaml.

10. [cmd_build_init_test.go] — Update test expectations to check for `config.yaml` instead of `pipeline.yml`.
    Verify: `go test ./... -run TestCmdAgentInit` passes.

11. [cmd_init_test.go] — Update test expectations to verify prefix is in config.yaml, not standalone file.
    Verify: `go test ./... -run TestCmdInit` passes.

12. [examples/*/] — Rename `pipeline.yml` to `config.yaml` in all example directories and add `prefix:` field where appropriate.
    Verify: examples load correctly with new config name.

13. [.ko/] — Rename `.ko/pipeline.yml` to `.ko/config.yaml` in the knockout project itself and add `prefix: ko` field.
    Verify: dogfooding project uses new config format.

14. [README.md or docs] — Update any documentation that references `pipeline.yml` to use `config.yaml`. Search for mentions of the old filename.
    Verify: docs reflect new config filename.

## Open Questions

**Q1: Should we keep backwards compatibility with `pipeline.yml`?**

The ticket says "consolidate" without mentioning migration. INVARIANTS.md says "avoid backwards-compatibility hacks" and "if you are certain that something is unused, you can delete it completely."

However, the ticket is a prereq for ko-1930 (agent harness extraction), suggesting this is infrastructure work. A hard break could disrupt users mid-development.

**Recommendation:** Add a deprecation path: `FindPipelineConfig` tries `config.yaml` first, falls back to `pipeline.yml` with a warning to stderr. Remove fallback in a future cleanup ticket. This is minimal compatibility, not a hack — just search order.

**Q2: What happens to the standalone `.ko/prefix` file for projects that have it?**

Old projects will have `.ko/prefix`. The code should prefer `config.yaml` if it exists, but fall back to `.ko/prefix` for projects not yet migrated.

**Recommendation:** `ReadPrefix` checks config first, then falls back to reading `.ko/prefix` file. `WritePrefix` only writes to config (no dual-write). Natural migration path: once config exists with prefix, old file is ignored.

**Q3: Should prefix be writable after init, or is it set-once?**

Currently `WritePrefix` exists and is called from `detectPrefix`. But prefix changes are rare/dangerous (breaks existing ticket IDs). Should the new config support runtime writes, or only be set during init?

**Recommendation:** Keep it writable for now (maintain existing behavior), but the implementation can be simple: read full config, update prefix field, write back. Don't optimize for a rare operation.
