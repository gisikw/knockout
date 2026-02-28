---
id: ko-dc03
status: closed
deps: []
created: 2026-02-27T15:15:31Z
type: task
priority: 2
---
# Agent loop should detect flake.nix and wrap harness invocations in 'nix develop --command' automatically, so pipelines don't need to handle nix environments themselves

## Notes

**2026-02-27 15:30:17 UTC:** ko: FAIL at node 'verify' — node 'verify' failed after 3 attempts: command failed: exit status 127
sh: line 1: if [ -f flake.nix ]; then nix develop --command just test; else just test; fi: command not found


**2026-02-27 15:45:14 UTC:** # Implementation Summary: Automatic Nix Environment Detection

## What Was Done

Successfully implemented automatic flake.nix detection in the agent loop, eliminating the need for pipelines to manually handle nix environment wrapping.

### Core Implementation

1. **Added `hasFlakeNix()` helper** (build.go:821-826)
   - Checks if `flake.nix` exists in project root by calling `ProjectRoot(ticketsDir)`
   - Returns boolean indicating presence

2. **Modified `runPromptNode()`** (build.go:488-500)
   - Detects flake.nix before creating context-aware command
   - If present, prepends `nix develop --command` to harness invocation args
   - Ensures all agent harness invocations run within nix environment when available

3. **Modified `runRunNode()`** (build.go:538-547)
   - Derives ticketsDir from artifactDir to check for flake.nix
   - Conditionally wraps shell commands: `nix develop --command sh -c <cmd>` when flake.nix exists
   - Maintains original `sh -c <cmd>` behavior when no flake.nix

4. **Simplified pipeline configuration** (.ko/pipeline.yml:47)
   - Removed manual flake.nix check: `if [ -f flake.nix ]; then nix develop --command just test; else just test; fi`
   - Replaced with simple `just test` - wrapping now happens automatically

5. **Cleaned up prompts** (.ko/prompts/)
   - Removed environment notes about tools not being on PATH from implement.md (lines 34-37)
   - Removed nix develop wrapper instructions from review.md (lines 33-36)
   - Agents no longer need to manually wrap commands

6. **Added test coverage** (build_test.go:241-260)
   - `TestHasFlakeNix` verifies detection logic in both cases (exists/doesn't exist)
   - All tests pass

## Notable Decisions

- **Placement of `hasFlakeNix()`**: Added after line 800 in build.go as planned, keeping helper functions grouped at the end of the file
- **TicketsDir derivation in `runRunNode()`**: Used `filepath.Dir(filepath.Dir(artifactDir))` to derive ticketsDir from artifactDir, matching the project's directory structure (.ko/tickets/id.artifacts/)
- **Command wrapping approach**: For prompt nodes, wrapped the args before creating CommandContext; for run nodes, conditionally created different command structures - both approaches preserve all existing environment variables and execution context

## Verification

- ✅ All planned tasks completed
- ✅ `go test ./...` passes (including new TestHasFlakeNix)
- ✅ Pipeline configuration simplified as intended
- ✅ No invariant violations:
  - Decision logic is pure (hasFlakeNix just checks file existence)
  - No runtime dependencies introduced (still just the binary)
  - Implementation follows existing patterns in build.go
- ✅ The fix addresses the root cause: the original error was "command not found" when running the complex shell conditional - now the conditional is gone and wrapping happens in Go code where it's properly executed

## Future Implications

This change makes all pipelines automatically nix-aware. Any project with a flake.nix at its root will now have all harness invocations and run commands automatically wrapped in `nix develop --command`, without requiring manual pipeline configuration.

**2026-02-27 15:45:14 UTC:** ko: SUCCEED
