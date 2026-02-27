## Goal
Automatically detect flake.nix in project root and wrap agent harness invocations in `nix develop --command`, removing the need for pipelines to manually handle nix environments.

## Context
The agent loop executes workflows by invoking agent harnesses (shell scripts in `agent-harnesses/`). Currently, built-in harnesses (`claude.sh`, `cursor.sh`) invoke the agent CLI directly. The project's own pipeline had a manual check for `flake.nix` in the verify node's `run` command (line 47 in `.ko/pipeline.yml`), but this has now been simplified to just `just test` in anticipation of automatic wrapping.

Key files:
- `harness.go:76` — `ShellAdapter.BuildCommand()` sets KO_* env vars and executes harness scripts
- `build.go:432` — `runPromptNode()` builds the command via adapter, then executes it with additional env vars
- `build.go:528` — `runRunNode()` executes shell commands with env vars (also needs wrapping)
- `ticket.go:549` — `ProjectRoot(ticketsDir string)` returns project root from tickets directory
- `agent-harnesses/claude.sh` and `cursor.sh` — built-in harness scripts that invoke CLIs

The recent commit (27d3e1b) added manual flake.nix detection to the pipeline config and documented the restriction in prompts. This ticket moves that logic into the agent loop itself, so it applies automatically to all prompt nodes and run nodes without manual checks.

The error that triggered this ticket shows that attempting to run the old complex shell command failed because it wasn't being properly executed through `sh -c`. The new approach both simplifies the pipeline config AND fixes this by handling nix wrapping at the build execution layer.

## Approach
Add a flake.nix detection helper function. In both `runPromptNode()` and `runRunNode()`, detect if `flake.nix` exists in the project root. If it does, wrap command execution by prepending `nix develop --command` to the command. For prompt nodes, this modifies the harness invocation arguments. For run nodes, this wraps the shell execution.

## Tasks
1. [build.go] — Add helper function `hasFlakeNix(ticketsDir string) bool` near the bottom of the file (after line 800). Function checks if `flake.nix` exists in `ProjectRoot(ticketsDir)` using `os.Stat`.
   Verify: Function returns true when flake.nix exists, false otherwise.

2. [build.go:runPromptNode] — After building the command via adapter (line 486), detect if flake.nix exists. If present, wrap the command by prepending `nix develop --command` to cmd.Args before creating the context-aware command. Store the wrapped args in a variable `cmdArgs` and use that when creating `cmdCtx`.
   Verify: When flake.nix exists, harness invocations include `nix develop --command` prefix.

3. [build.go:runRunNode] — After function signature (line 528), detect if flake.nix exists by deriving ticketsDir from artifactDir (`filepath.Dir(filepath.Dir(artifactDir))`). When creating the command (line 532), conditionally wrap: if flake.nix exists, use `nix develop --command sh -c <node.Run>`, otherwise use `sh -c <node.Run>` as currently.
   Verify: When flake.nix exists, run commands are wrapped in nix develop.

4. [.ko/pipeline.yml:47] — ✅ DONE. The manual flake.nix check has been removed and replaced with `just test`.
   Verify: Pipeline config is cleaner, wrapping will happen automatically once build.go changes are in place.

5. [.ko/prompts/implement.md:34-36, .ko/prompts/review.md:33-35] — Remove environment notes about tools not being on PATH and needing nix develop wrappers (added in commit 27d3e1b). These restrictions no longer apply since wrapping will be automatic.
   Verify: Prompts are cleaner, no manual nix instructions.

6. [build_test.go] — Add test function `TestHasFlakeNix` that creates a temporary directory structure with tickets dir, creates/doesn't create a flake.nix file in the project root, and verifies the function returns the correct boolean. Follow existing test patterns in the file.
   Verify: `go test ./...` passes with new test.

## Open Questions
None. The implementation is straightforward: detect flake.nix at project root, wrap command execution conditionally. The pipeline.yml has already been simplified in preparation. We're implementing the build execution layer changes that make this work automatically.
