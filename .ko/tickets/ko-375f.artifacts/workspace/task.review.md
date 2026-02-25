Perfect! The implementation is complete and correct. Let me write my final disposition:

---

## Review Complete: Implementation Successful ✅

I've thoroughly reviewed the implementation against the plan, ticket requirements, and project invariants. The migration from YAML to shell-based agent harnesses was executed successfully.

### Compliance Summary

**✅ Plan Compliance**: All 17 planned tasks completed
- Created spec file (`specs/agent_harnesses.feature`)
- Created shell harnesses (`claude.sh`, `cursor.sh`)
- Implemented `ShellAdapter` 
- Removed all YAML support (`TemplateAdapter`, YAML parsing)
- Updated tests and documentation

**✅ Correctness**: Implementation matches ticket requirements
- Shell scripts receive KO_* environment variables
- Binary fallback moved to shell (cursor.sh uses `command -v`)
- Prompt passing via KO_PROMPT environment variable
- Both stdin (claude) and argument (cursor) prompt methods supported

**✅ Completeness**: All components present
- Specs, implementation, tests, and documentation all updated
- Edge cases handled (empty model, no system prompt, no allowed tools)

**✅ Invariants**: All project rules followed
- Spec created before/alongside code
- File sizes under 500 lines (harness.go: 98 lines, down from 173)
- Zero external runtime dependencies (shell is acceptable per INVARIANTS.md)
- Tests passing for all changes

### Issues Found and Fixed

**Issue**: Cursor.sh had improper shell argument quoting
- **Risk**: Prompts with spaces would break
- **Fix**: Moved prompt to separate quoted argument: `exec $BINARY -p "$COMBINED_PROMPT" $args`
- **Status**: Fixed and verified with tests

**Issue**: build.go was overwriting adapter environment variables
- **Risk**: KO_* variables wouldn't reach shell scripts
- **Fix**: Preserve `cmd.Env` when set by adapter
- **Status**: Fixed during implementation, documented in summary

### Key Improvements

1. **Code Simplification**: 47% reduction in harness.go (173→98 lines)
2. **Architecture**: Cleaner separation - shell scripts control invocation, Go just sets env vars
3. **Flexibility**: Each harness can handle binary fallback and argument construction independently
4. **Maintainability**: Adding new harnesses is now just creating a shell script

All tests pass. Documentation updated. No security issues. Ready to proceed.

```json
{"disposition": "continue"}
```
