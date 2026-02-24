Perfect! Let me now provide my final disposition. The implementation is complete, correct, and compliant with all requirements.

## Review Summary

The implementation successfully extracted hardcoded agent adapters into declarative YAML configuration files. All planned tasks were completed:

✅ **Plan Compliance**: All 6 tasks completed
- Built-in harness YAML files created (claude.yaml, cursor.yaml)
- Harness loading infrastructure implemented (harness.go, 167 lines)
- Adapter.go refactored to use harness loader (75 lines removed)
- Comprehensive tests added (harness_test.go, 263 lines, all passing)
- go:embed used to ship harnesses with binary
- Documentation updated (README.md, specs/pipeline.feature)

✅ **Correctness**: Implementation matches ticket requirements
- Agents are now extensible without recompiling
- Built-in harnesses (claude, cursor) ship embedded in binary
- User-supplied harnesses supported in `~/.config/knockout/agent-harnesses/`
- Project-local harnesses supported in `.ko/agent-harnesses/`
- Backward compatibility maintained (RawCommandAdapter preserved)

✅ **INVARIANTS.md Compliance**
- File size limits respected (167 and 263 lines, under 500)
- No util.go created
- Decision logic separated from I/O
- Tests mirror source files

✅ **Security**: No vulnerabilities
- Template variables passed as separate args (no injection)
- Safe path construction with filepath.Join
- No arbitrary file read outside harness directories

✅ **Completeness**: All edge cases handled
- Conditional flag rendering works correctly
- stdin vs arg-based prompt passing handled
- Binary fallback resolution implemented
- Empty/missing values omitted appropriately

✅ **Tests**: All tests pass (9/9 harness tests, plus all existing tests)

**Minor Note**: The plan recommended simple empty-string omission for conditional flags (Q1), but the implementation uses a more sophisticated newline-separated multi-arg expansion approach. This deviation is justified and results in cleaner template logic.

```json
{"disposition": "continue"}
```
