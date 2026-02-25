## Goal
Consolidate `.ko/pipeline.yml` and `.ko/prefix` into a single `.ko/config.yaml` file.

## Context
Currently, the `.ko/` directory has multiple top-level config files:
- `pipeline.yml` - contains pipeline config (workflows, hooks, model settings, etc.)
- `prefix` - single-line file containing the ticket ID prefix (e.g., "ko")
- Runtime state files: `agent.lock`, `agent.pid`, `agent.log`, `agent.heartbeat` (already gitignored)

Key files involved:
- `pipeline.go` - defines `Pipeline` struct and parsing logic with `FindPipelineConfig()`, `LoadPipeline()`, `ParsePipeline()`
- `cmd_create.go` - contains `ReadPrefix()`, `WritePrefix()`, `detectPrefix()` functions
- `cmd_init.go` - writes the prefix file via `WritePrefix()`
- `cmd_build_init.go` - scaffolds pipeline.yml via `cmdAgentInit()`
- `harness.go` - agent harness search looks in `.ko/agent-harnesses/`
- `README.md` - documents pipeline configuration and structure

The ticket mentions this is a prereq for agent harness extraction (ko-1930), indicating we need to establish the config surface before deciding where harness overrides live.

Testing patterns: The codebase uses testscript (.txtar files) for integration tests, with corresponding specs in specs/ directory. Tests exist for both pipeline parsing (`pipeline_test.go`) and ticket creation with prefix detection (`testdata/ticket_creation/`).

## Approach
Create a unified config schema that combines pipeline settings and project-level metadata (prefix, etc.) in `.ko/config.yaml`. The new format will have top-level sections: `project` (for prefix and other project settings) and `pipeline` (containing current pipeline.yml content). Support backwards compatibility by reading from `pipeline.yml` + `prefix` if `config.yaml` doesn't exist, but all new scaffolding uses the unified format.

## Tasks
1. [pipeline.go] - Add new `Config` struct wrapping `Pipeline` and project settings (prefix). Add `LoadConfig()` function that tries `config.yaml` first, falls back to loading `pipeline.yml` + `prefix` separately for backwards compatibility. Update `FindPipelineConfig()` to `FindConfig()` that searches for either file.
   Verify: Unit tests pass for parsing both new and legacy formats.

2. [pipeline.go:ParsePipeline] - Refactor to extract YAML parsing utilities into reusable functions if needed. Add `ParseConfig()` function for new unified format that handles nested `pipeline:` and `project:` sections.
   Verify: Can parse both flat pipeline.yml and nested config.yaml formats.

3. [cmd_create.go] - Update `ReadPrefix()` to try reading from `config.yaml` first (under `project.prefix`), fall back to `.ko/prefix` file for backwards compatibility. Mark `WritePrefix()` as deprecated but keep functional.
   Verify: Existing tests for prefix detection still pass.

4. [cmd_build_init.go] - Update `cmdAgentInit()` to scaffold `.ko/config.yaml` instead of `.ko/pipeline.yml`. Update default template constant to use unified format with `project:` and `pipeline:` sections.
   Verify: `ko agent init` creates config.yaml with correct structure.

5. [cmd_init.go] - Update `cmdInit()` to write prefix to `.ko/config.yaml` (creating a minimal config with just the project section) instead of `.ko/prefix` file.
   Verify: `ko init <prefix>` creates config.yaml, not separate prefix file.

6. [build.go] - Update all calls to `FindPipelineConfig()`/`LoadPipeline()` to use new `FindConfig()`/`LoadConfig()` functions. Extract pipeline from the wrapper config struct.
   Verify: Build commands work with both old and new config formats.

7. [pipeline_test.go] - Add tests for unified config format parsing. Test backwards compatibility fallback paths.
   Verify: All new tests pass, existing pipeline tests continue to pass.

8. [testdata/pipeline/build_init.txtar] - Update test expectations to check for `.ko/config.yaml` instead of `.ko/pipeline.yml`.
   Verify: Test passes with new scaffold format.

9. [README.md] - Update documentation to reference `.ko/config.yaml` as the config file. Document the structure with `project:` and `pipeline:` sections. Add note about backwards compatibility with `pipeline.yml` + `prefix` files.
   Verify: Documentation accurately reflects new behavior.

## Open Questions
1. Should we include a migration command (`ko migrate-config`) to help users convert existing setups, or is transparent fallback sufficient?
   - Decision: Start with transparent fallback only. Users naturally migrate when they run `ko agent init` in a new project or manually update their config. A migration command adds complexity that may not be needed.

2. What other project-level settings should go in the `project:` section beyond `prefix`?
   - The ticket mentions "default priority" and "agent harness overrides" as candidates
   - Decision: Start with just `prefix` in the `project:` section for this ticket. Default priority and harness overrides can be added in follow-up tickets as their requirements become clear. This keeps the change focused and establishes the config structure without over-engineering.

3. Should `config.yaml` completely replace `pipeline.yml` or can they coexist?
   - Decision: They can coexist for backwards compatibility. The loader checks for `config.yaml` first, falls back to `pipeline.yml` + `prefix` if not found. This allows gradual migration without breaking existing setups.

4. Should the `.ko/.gitignore` be updated?
   - The `.gitignore` currently excludes runtime state files (agent.lock, agent.pid, agent.log, agent.heartbeat)
   - Decision: No changes needed. Config files (both config.yaml and legacy pipeline.yml) should be committed to version control.
