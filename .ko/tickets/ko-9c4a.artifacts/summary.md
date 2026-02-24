# Implementation Summary: ko-9c4a

## What Was Done

Successfully implemented the `plan-questions` frontmatter schema and `ko block --questions` flag as specified in the ticket.

### Changes Made

1. **ticket.go** (lines 19-51):
   - Added `PlanQuestions` field to `Ticket` struct with `yaml:"plan-questions,omitempty"` tag
   - Defined `PlanQuestion` struct with fields: `ID`, `Question`, `Context` (optional), `Options`
   - Defined `QuestionOption` struct with fields: `Label`, `Value`, `Description` (optional)

2. **ticket.go** (lines 111-127):
   - Extended `FormatTicket()` to serialize `PlanQuestions` to YAML with proper indentation (2-space indent per level)
   - Uses quoted strings for question/label/description, unquoted for value

3. **ticket.go** (lines 154-296):
   - Extended `ParseTicket()` with indentation-based state machine to parse `plan-questions:` from frontmatter
   - State tracking: `inPlanQuestions`, `currentQuestion`, `currentOption`, `inOptions`
   - Properly handles indent levels (0 for top-level keys, 2 for question items, 4 for properties, 6 for option items, 8 for option properties)
   - Reuses existing `countIndent()` helper from pipeline.go
   - Added `unquote()` helper to strip surrounding quotes from parsed strings

4. **cmd_status.go** (lines 80-103):
   - Added `ValidatePlanQuestions()` pure validation function
   - Validates required fields: `id`, `question`, `options` (must have at least one)
   - For each option: validates `label` and `value` are present
   - Returns clear error messages with context (question index, question ID, option index)

5. **cmd_status.go** (lines 105-179):
   - Replaced simple `cmdBlock()` delegation with full implementation
   - Added flag parsing for `--questions <json>`
   - Parses JSON into `[]PlanQuestion`, validates structure
   - If `--questions` provided: parses, validates, writes to ticket frontmatter
   - If `--questions` omitted: behaves as before (just sets status to blocked)
   - Emits mutation event with old/new status

6. **cmd_status_test.go**:
   - Added `TestValidatePlanQuestions()` with 10 test cases covering valid/invalid inputs
   - Tests empty slice, minimal fields, all fields, multiple questions, and all validation error cases

7. **ticket_test.go** (lines 119-458):
   - Added `TestPlanQuestions()` with 4 test cases: empty array, single question, question with context/descriptions, multiple questions
   - Added `TestPlanQuestionsRoundTrip()` to verify parse → format → parse produces identical struct
   - All tests verify field-by-field comparison

8. **specs/ticket_status.feature** (lines 52-92):
   - Added 8 gherkin scenarios for `ko block` behavior:
     - Block with valid questions JSON
     - Block without questions flag (backward compatibility)
     - Block with invalid JSON
     - Block with questions missing each required field (id, question, options)
     - Block with option missing label
     - Block with option missing value

## Notable Decisions

1. **Indentation parsing**: Used the same indentation-based state machine pattern established in `pipeline.go`. This maintains consistency with the codebase's minimal YAML parser approach (no external dependencies).

2. **Quote handling**: Question text and option labels/descriptions are serialized with quotes and parsed with an `unquote()` helper. Option values are left unquoted to match the YAML examples in the ticket schema.

3. **Validation placement**: Created a pure `ValidatePlanQuestions()` function separate from `cmdBlock()`, following the INVARIANTS.md principle that "decision logic is pure."

4. **Backward compatibility**: Plain `ko block <id>` without `--questions` continues to work exactly as before, just setting status to blocked.

5. **Event emission**: Emits a mutation event when blocking, consistent with other status change commands.

## Test Coverage

- All Go tests pass (validation tests, parsing tests, round-trip tests)
- All gherkin scenarios added to `specs/ticket_status.feature`
- Manual verification: `ko block ko-9c4a --questions '[...]'` correctly writes questions to frontmatter and sets status

## Compliance

- ✅ INVARIANTS.md: Every behavior has a spec (added 8 scenarios to ticket_status.feature)
- ✅ INVARIANTS.md: Decision logic is pure (ValidatePlanQuestions is pure)
- ✅ INVARIANTS.md: Fail fast on bad input (validation errors are immediate)
- ✅ INVARIANTS.md: File size limit (all modified files under 500 lines)
- ✅ Plan compliance: All 7 planned tasks completed
- ✅ No unexplained deviations from plan
- ✅ Scope maintained: only added the requested feature, no extra refactoring

The implementation is complete and ready for use.
