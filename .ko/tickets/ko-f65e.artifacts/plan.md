## Goal
Consolidate `.ko/pipeline.yml` and other top-level config files into a single `.ko/config.yaml` that serves as the project-level configuration surface.

## Context
Currently `.ko/` has several top-level files:
- `pipeline.yml` — pipeline configuration (workflows, model defaults, hooks)
- `prefix` — project ticket prefix (e.g., "ko")
- `agent.lock` — runtime lock file (gitignored)
- `agent.pid` — runtime PID file (gitignored)
- `agent.log` — runtime log file (gitignored)
- `agent.heartbeat` — runtime heartbeat file (gitignored)
- `prompts/` — prompt templates directory

The runtime files (lock, pid, log, heartbeat) are ephemeral and should remain separate. The prompts directory should remain a directory.

The consolidation target is:
- `pipeline.yml` content → `config.yaml` under a `pipeline:` section
- `prefix` content → `config.yaml` under a `prefix:` field

This creates a clear config surface before ko-1930 (agent harness extraction), which asks where project-level harness overrides should live — the answer will be in `config.yaml`.

### Key files
- `pipeline.go` (lines 45-62): `FindPipelineConfig()` and `LoadPipeline()` — hardcoded to look for `pipeline.yml`
- `cmd_create.go` (lines 183-202): `ReadPrefix()` and `WritePrefix()` — read/write `.ko/prefix`
- `cmd_init.go` (lines 31-33, 43-46): uses `ReadPrefix()` and `WritePrefix()`
- `cmd_agent.go` (line 217, 257): calls `FindPipelineConfig()`
- `cmd_build.go` (line 55, 61): calls `FindPipelineConfig()` and `LoadPipeline()`
- `cmd_loop.go` (line 114, 120): calls `FindPipelineConfig()` and `LoadPipeline()`

### Parsing approach
The codebase uses a custom minimal YAML parser (no external deps) — see `pipeline.go:ParsePipeline()`. The same approach should be used for the new config format.

### Migration strategy
Both old and new formats should be supported during a transition period:
1. If `config.yaml` exists, use it
2. If `config.yaml` doesn't exist but `pipeline.yml` exists, use `pipeline.yml` + `prefix` file
3. For writes (prefix), update `config.yaml` if it exists, otherwise fall back to `prefix` file

## Approach
Create a new unified config structure that wraps the existing Pipeline struct and adds a prefix field. Update all config discovery to look for `config.yaml` first, falling back to `pipeline.yml` for backward compatibility. Update prefix read/write functions to prefer `config.yaml` when it exists. Update `ko agent init` to generate `config.yaml` instead of `pipeline.yml`.

## Tasks
1. [pipeline.go:Pipeline] — Create a new `ProjectConfig` struct containing `Prefix string` and `Pipeline *Pipeline` fields. Add `ParseProjectConfig(content string)` function that parses the unified YAML format with `prefix:` at top level and `pipeline:` as a nested section. Use the same minimal YAML parsing approach as `ParsePipeline()`.
   Verify: `go test ./... -run TestParseProjectConfig` passes (will add test in step 2).

2. [pipeline_test.go] — Add tests for `ParseProjectConfig()`: test parsing unified config with both prefix and pipeline sections, test that pipeline section matches existing pipeline.yml format, test backward compatibility with missing prefix field, test validation errors.
   Verify: `go test ./... -run TestParseProjectConfig` passes.

3. [pipeline.go:FindPipelineConfig] — Rename to `FindProjectConfig()` and update to look for `.ko/config.yaml` first, falling back to `.ko/pipeline.yml` if config.yaml doesn't exist. Return both the path and a boolean indicating which format was found.
   Verify: `go test ./... -run TestFindProjectConfig` passes.

4. [pipeline.go:LoadPipeline] — Update to call `FindProjectConfig()` internally and handle both formats. If loading from `config.yaml`, call `ParseProjectConfig()` and return the embedded Pipeline. If loading from `pipeline.yml`, call `ParsePipeline()` directly.
   Verify: existing pipeline tests still pass.

5. [cmd_create.go:ReadPrefix,WritePrefix] — Update both functions to check for `config.yaml` first. `ReadPrefix()` should try parsing `config.yaml` for the prefix field, falling back to the `prefix` file. `WritePrefix()` should update `config.yaml` if it exists, otherwise fall back to writing the `prefix` file.
   Verify: `go test ./... -run TestReadWritePrefix` passes.

6. [cmd_build_init.go] — Update `cmdAgentInit()` to generate `.ko/config.yaml` instead of `.ko/pipeline.yml`. The generated config should include a commented-out `prefix:` field as an example. Update the success message to reference `config.yaml`.
   Verify: `go test ./... -run TestCmdBuildInit` passes and generates config.yaml.

7. [examples/*/pipeline.yml] — For each example directory (default, minimal, structured), rename `pipeline.yml` to `config.yaml` and wrap the content under a `pipeline:` section (indent all lines by 2 spaces and add `pipeline:` header).
   Verify: examples parse correctly via manual test with `ko agent build`.

8. [.ko/pipeline.yml] — Rename knockout's own `.ko/pipeline.yml` to `.ko/config.yaml`, wrap content under `pipeline:` section, and add `prefix: ko` at the top.
   Verify: `ko ready` and `ko agent status` work correctly.

9. [README.md] — Update all references from `.ko/pipeline.yml` to `.ko/config.yaml`. Update the pipeline configuration section to show the new unified format with `prefix:` and `pipeline:` sections. Add a note about backward compatibility with `pipeline.yml`.
   Verify: documentation accurately describes the new config format.

## Open Questions
1. Should we add a `default_priority` field to config.yaml now (mentioned in ticket), or defer that to a separate ticket since it's not implemented yet?

2. Should the migration period be documented with a deprecation timeline for pipeline.yml, or should we support both formats indefinitely? (Given the prereq relationship to ko-1930, likely needs to be stable before that work begins.)

3. Should `ko init` be updated to write config.yaml with prefix, or should that remain separate since `ko init <prefix>` only sets up the tickets directory and prefix (and `ko agent init` sets up the pipeline)?
