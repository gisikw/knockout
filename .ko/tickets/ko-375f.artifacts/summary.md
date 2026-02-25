# Implementation Summary: Shell-Based Agent Harnesses

## What Was Done

Successfully migrated agent harnesses from YAML templates to executable shell scripts. The implementation completely replaced the YAML-based `TemplateAdapter` system with a simpler `ShellAdapter` that passes parameters via KO_-namespaced environment variables.

## Key Changes

### Architecture
- **Removed**: YAML parsing, template rendering, `TemplateAdapter` with 100+ lines of template logic
- **Added**: `ShellAdapter` (20 lines) that simply sets environment variables and executes scripts
- **Result**: Simpler, more flexible architecture where harness scripts have full control over agent invocation

### Built-in Harnesses
- `agent-harnesses/claude.sh`: Passes prompt via stdin using `echo "$KO_PROMPT" | claude ...`
- `agent-harnesses/cursor.sh`: Implements binary fallback (`command -v cursor-agent || command -v agent`) and combines system prompt with user prompt (cursor doesn't support separate system prompts)
- Both scripts handle conditional flags based on KO_* environment variables

### Environment Variable Contract
Shell harnesses receive:
- `KO_PROMPT`: Full prompt text
- `KO_MODEL`: Model name (may be empty)
- `KO_SYSTEM_PROMPT`: System prompt (may be empty)
- `KO_ALLOW_ALL`: "true" or "false" string
- `KO_ALLOWED_TOOLS`: Comma-separated list of tool names (may be empty)

### Code Simplification
- `harness.go`: Reduced from 173 lines to 92 lines (47% reduction)
- Removed entire `TemplateAdapter` implementation (binary resolution, template variable expansion, stdin setup)
- Removed `yaml.v3` dependency from harness loading
- `HarnessConfig` simplified to just contain `ScriptPath`

### Testing
- Added `specs/agent_harnesses.feature` with comprehensive behavioral scenarios
- Added `testdata/agent_harnesses/shell_harness.txtar` for end-to-end testing
- Updated all existing harness tests to work with shell scripts
- All tests pass

## Notable Implementation Decisions

### 1. Build.go Environment Variable Preservation
**Issue**: `build.go:runPromptNode` was replacing `cmd.Env` with `os.Environ()`, losing the KO_* variables set by `ShellAdapter`.

**Fix**: Updated to check if `cmd.Env` is already set by the adapter and preserve it:
```go
if cmd.Env != nil {
    cmdCtx.Env = append(cmd.Env, /* workspace vars */)
} else {
    cmdCtx.Env = append(os.Environ(), /* workspace vars */)
}
```

This ensures adapter-set environment variables flow through to the executed harness script.

### 2. Cursor Prompt Argument Quoting
**Issue**: Original implementation built prompt into args string: `args="-p $COMBINED_PROMPT ..."`, which would break on prompts with spaces.

**Fix**: Pass prompt as separate quoted argument: `exec $BINARY -p "$COMBINED_PROMPT" $args`

This ensures prompts with special characters are properly handled by the shell.

### 3. Built-in Harness Extraction
Built-in shell scripts are embedded via `//go:embed agent-harnesses/*.sh` and extracted to temp files with 0755 permissions when loaded. This allows the Go code to execute them like user-provided scripts.

## Migration Path for Users

Users with custom YAML harnesses must convert them to shell scripts:

**Before (YAML):**
```yaml
binary: mycli
args:
  - "-p"
  - "${prompt}"
  - "${model}"
```

**After (Shell):**
```bash
#!/bin/sh
set -e
args=""
if [ -n "$KO_MODEL" ]; then
  args="$args --model $KO_MODEL"
fi
echo "$KO_PROMPT" | mycli $args
```

## Invariants Compliance

- ✅ **Specs before code**: Created `specs/agent_harnesses.feature`
- ✅ **Every spec has a test**: `testdata/agent_harnesses/shell_harness.txtar`
- ✅ **500-line file limit**: `harness.go` reduced to 92 lines
- ✅ **Zero external runtime dependencies**: Shell scripts use only POSIX shell built-ins (`command -v`, `echo`, `exec`, `if`, `set`)
- ✅ **No YAML runtime dependency**: Removed `yaml.v3` from harness code path

## Future Maintenance Notes

1. **Adding New Built-in Harnesses**: Create a `.sh` file in `agent-harnesses/`, ensure it's executable, and the embed directive will automatically include it.

2. **User Harnesses**: Must be placed at `.ko/agent-harnesses/<name>` (project) or `~/.config/knockout/agent-harnesses/<name>` (user) with executable permissions. No file extension.

3. **Environment Variable Changes**: If new parameters need to be passed to harnesses, add them to `ShellAdapter.BuildCommand()` as `KO_*` variables. Update the spec file to document the new variable.

4. **Backward Compatibility**: YAML support has been completely removed. Users must migrate to shell scripts.
