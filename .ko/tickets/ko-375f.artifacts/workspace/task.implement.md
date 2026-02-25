Perfect! The implementation is complete. Let me provide a summary:

## Summary

I have successfully migrated agent harnesses from YAML templates to shell scripts. Here's what was changed, organized by task:

### Task 1: Created spec file
- Added `specs/agent_harnesses.feature` documenting shell harness behavior, environment variable contract, search order, and migration from YAML

### Tasks 2-3: Created shell harnesses
- `agent-harnesses/claude.sh`: Shell script for Claude CLI that receives KO_* env vars and passes prompt via stdin
- `agent-harnesses/cursor.sh`: Shell script for Cursor CLI with binary fallback logic and combined system+user prompt

### Task 4: Added ShellAdapter
- Created `ShellAdapter` struct in `harness.go` that implements `AgentAdapter`
- Sets KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, and KO_ALLOWED_TOOLS environment variables

### Tasks 5-6: Refactored harness loading
- Simplified `HarnessConfig` struct to only contain `ScriptPath` (shell-only)
- Updated `LoadHarness` to search for executable shell scripts (no .sh extension for user/project, .sh for built-ins)
- Embedded built-ins are extracted to temp files with exec permissions

### Task 7: Updated LookupAdapter
- Modified `adapter.go` to return `ShellAdapter` directly (removed YAML path)

### Task 8: Added ShellAdapter tests
- Tests verify KO_* env vars are set correctly
- Tests verify shell harnesses work for built-in, user, and project locations

### Task 9: Added end-to-end testscript test
- `testdata/agent_harnesses/shell_harness.txtar` verifies shell harness receives correct env vars during a build
- Fixed `build.go` to preserve `cmd.Env` from adapters when creating context-aware commands

### Tasks 10-15: Removed YAML support
- Deleted `agent-harnesses/claude.yaml` and `cursor.yaml`
- Removed `parseYAMLHarness`, `TemplateAdapter`, and all YAML-related code
- Removed yaml.v3 import
- Updated embed directive to only include `*.sh` files
- Removed all YAML-specific tests

### Task 16: Updated documentation
- Rewrote README.md Custom Agent Harnesses section with shell script examples
- Documented KO_* environment variables
- Provided examples for binary fallback logic

### Task 17: Verified parameter passing
- Confirmed `allowedTools` is correctly passed through call chain: build loop → runNode → runPromptNode → BuildCommand

All tests pass, the code compiles successfully, and the migration is complete with YAML support fully removed.
