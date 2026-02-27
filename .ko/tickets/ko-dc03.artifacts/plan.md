## Goal
Automatically detect flake.nix in project root and wrap agent harness invocations in `nix develop --command`, removing the need for pipelines to manually handle nix environments.

## Context
The agent loop executes workflows by invoking agent harnesses (shell scripts in `agent-harnesses/`). Currently, built-in harnesses (`claude.sh`, `cursor.sh`) invoke the agent CLI directly. The project's own pipeline manually checks for `flake.nix` in the verify node's `run` command (line 47 in `.ko/pipeline.yml`), wrapping `just test` with `nix develop --command` when present.

Key files:
- `harness.go:76` — `ShellAdapter.BuildCommand()` sets KO_* env vars and executes harness scripts
- `build.go:431-525` — `runPromptNode()` builds the command via adapter, then executes it with additional env vars
- `build.go:527-551` — `runRunNode()` executes shell commands with env vars (also needs wrapping)
- `agent-harnesses/claude.sh` and `cursor.sh` — built-in harness scripts that invoke CLIs

The recent commit (27d3e1b) added manual flake.nix detection to the pipeline config and documented the restriction in prompts. This ticket moves that logic into the agent loop itself, so it applies automatically to all prompt nodes and run nodes without manual checks.

## Approach
Add a flake.nix detection function at project root level. In `runPromptNode()` and `runRunNode()`, after building the command context, detect if `flake.nix` exists in the project root. If it does, wrap the command execution by prepending `nix develop --command` to the exec.Cmd. This applies transparently to all harness invocations and run commands.

## Tasks
1. [build.go] — Add helper function `hasFlakeNix(ticketsDir string) bool` that checks if `flake.nix` exists in `ProjectRoot(ticketsDir)`.
   Verify: Function returns true when flake.nix exists, false otherwise.

2. [build.go:runPromptNode] — After building the command via adapter (line 486) and before creating the context-aware command (line 489), detect flake.nix. If present, wrap the command by modifying `cmd` to prepend `nix develop --command` before the agent harness invocation.
   Verify: Harness invocations are wrapped when flake.nix exists.

3. [build.go:runRunNode] — After creating the shell command (line 532), apply the same flake.nix detection and wrapping logic.
   Verify: Shell run commands are wrapped when flake.nix exists.

4. [.ko/pipeline.yml:47] — Remove manual flake.nix check from verify node's run command. Replace `"if [ -f flake.nix ]; then nix develop --command just test; else just test; fi"` with `just test`.
   Verify: Pipeline config is cleaner, wrapping happens automatically.

5. [.ko/prompts/implement.md, .ko/prompts/review.md] — Remove environment notes about tools not being on PATH and needing nix develop wrappers (added in commit 27d3e1b). These restrictions no longer apply since wrapping is automatic.
   Verify: Prompts are cleaner, no manual nix instructions.

6. [build_test.go or new file] — Add test case that verifies command wrapping when flake.nix is present in project root.
   Verify: `go test ./...` passes with new test.

## Open Questions
None. The implementation is straightforward: detect flake.nix at project root, wrap command execution conditionally. The existing pipeline already validates this approach works (commit 27d3e1b shows it in production). We're just moving the logic from pipeline config into the build execution layer.
