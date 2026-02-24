## Review Complete

I've reviewed the implementation against the plan and ticket requirements. Here's my analysis:

### ✅ Plan Compliance
All 7 planned tasks were completed:
1. ✅ Added `PlanQuestions` field to `Ticket` struct with proper type definitions
2. ✅ Extended `FormatTicket()` to serialize with proper YAML indentation
3. ✅ Extended `ParseTicket()` with state machine parsing
4. ✅ Added comprehensive table-driven tests for parsing and formatting
5. ✅ Replaced `cmdBlock()` with full implementation including `--questions` flag
6. ✅ Added pure `ValidatePlanQuestions()` validation function
7. ✅ Added 8 gherkin scenarios to `specs/ticket_status.feature`

### ✅ Correctness
- The implementation matches the ticket's requested schema exactly
- JSON-to-YAML conversion works correctly
- Backward compatibility preserved (plain `ko block` still works)
- All validation errors have clear, helpful messages

### ✅ Test Coverage
- 10 validation test cases covering all error conditions
- 4 parsing test cases with various field combinations
- Round-trip test verifying parse → format → parse integrity
- All tests pass

### ✅ INVARIANTS.md Compliance
- Every behavior has a spec in `specs/ticket_status.feature`
- Decision logic is pure (`ValidatePlanQuestions` has no I/O)
- Fail fast on bad input (validation errors are immediate)
- All modified files under 500 lines
- No multi-purpose functions

### ✅ Code Quality
- Follows existing patterns (indentation-based parsing like `pipeline.go`)
- Clear separation between decision logic and I/O
- Consistent with codebase style
- No unnecessary abstractions or over-engineering

### Manual Testing
Verified the feature works end-to-end: `ko block ko-9c4a --questions '[...]'` correctly writes questions to frontmatter with proper YAML formatting.

```json
{"disposition": "continue"}
```
