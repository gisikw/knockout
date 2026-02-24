Now let me verify there are no invariant violations and complete my review:

## Review Complete

The implementation successfully completes all planned tasks and passes all tests. Here's my assessment:

### ✅ Plan Compliance
All 4 planned tasks completed:
1. ✅ Specs added to `specs/ticket_notes.feature` 
2. ✅ Tests added to `testdata/ticket_notes/notes_basic.txtar`
3. ✅ Code modified in `cmd_note.go` with stdin detection
4. ✅ Tests pass

### ✅ Correctness
- Implements exactly what the ticket requested: multiline text via heredoc/stdin
- Maintains backward compatibility for command-line args
- Follows standard Unix conventions (stdin precedence)

### ✅ Completeness
- Edge cases handled: empty stdin, no args, stdin read errors
- Proper error messages for all failure modes
- Multiline formatting preserved correctly

### ✅ Safety
- No security issues (no injection, safe string handling)
- No unintended side effects
- Proper error handling throughout

### ✅ INVARIANTS.md Compliance
- Spec before/alongside code ✓
- Every spec has a test ✓
- Zero external runtime dependencies ✓
- Proper error handling to stderr with exit codes ✓

### ✅ Tests
All tests pass, including 4 new test scenarios covering stdin input patterns.

### ✅ Scope
Changes stayed strictly within scope. No unrelated modifications.

The implementation is clean, correct, and ready to merge.

```json
{"disposition": "continue"}
```
