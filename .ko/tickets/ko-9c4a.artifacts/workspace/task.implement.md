Perfect! All tests pass. Let me create a summary of what was implemented:

## Summary

I successfully implemented all tasks from the plan:

### Task 1: Add PlanQuestions field to Ticket struct ✅
- Added `PlanQuestions []PlanQuestion` field to `Ticket` struct with yaml tag `plan-questions,omitempty`
- Defined `PlanQuestion` struct with fields: `ID`, `Question`, `Context`, `Options`
- Defined `QuestionOption` struct with fields: `Label`, `Value`, `Description`
- **File**: `ticket.go:18-46`

### Task 2: Extend FormatTicket to serialize PlanQuestions ✅
- Extended `FormatTicket()` to write `plan-questions` as indented YAML (2-space indent per level)
- Questions use array notation with `- ` prefix
- Handles optional fields (`context`, `description`) correctly
- **File**: `ticket.go:95-116`

### Task 3: Extend ParseTicket to parse plan-questions ✅
- Implemented indentation-based state machine parser for `plan-questions` section
- Tracks parsing state through nested structures (questions → options)
- Uses existing `countIndent()` helper from `pipeline.go`
- Added `unquote()` helper to strip surrounding quotes from YAML string values
- **File**: `ticket.go:119-297`

### Task 4: Add table-driven tests ✅
- Created `TestPlanQuestions()` with 4 test cases:
  - Empty questions array
  - Single question with two options
  - Question with context and descriptions
  - Multiple questions
- Created `TestPlanQuestionsRoundTrip()` to verify parse → format → parse consistency
- All tests pass
- **File**: `ticket_test.go:121-391`

### Task 5: Update cmdBlock to handle --questions flag ✅
- Replaced simple delegation to `cmdStatus` with full implementation
- Added flag parsing for `--questions` with JSON string parameter
- Parses JSON into `[]PlanQuestion` and validates before writing
- Maintains backward compatibility - plain `ko block <id>` works without questions
- **File**: `cmd_status.go:105-181`

### Task 6: Add ValidatePlanQuestions validation function ✅
- Pure validation function checking all required fields:
  - Question must have `id`, `question`, and at least one option
  - Each option must have `label` and `value`
- Returns clear error messages with question/option indices
- Added comprehensive unit tests with 10 test cases
- **Files**: `cmd_status.go:80-103`, `cmd_status_test.go:1-166`

### Task 7: Add gherkin scenarios ✅
- Added 9 new scenarios to `specs/ticket_status.feature`:
  - Block with valid questions JSON
  - Block without questions flag
  - Block with invalid JSON
  - Block with missing required fields (id, question, options)
  - Block with option missing label/value
- Scenarios describe expected behavior without implementation details
- **File**: `specs/ticket_status.feature:52-93`

### Verification
- All Go tests pass: `go test ./...` ✅
- Build succeeds: `go build` ✅
- Manual testing confirms:
  - `ko block <id>` works without --questions ✅
  - `ko block <id> --questions '<json>'` writes questions to frontmatter ✅
  - Questions are properly formatted in YAML ✅
  - Questions are correctly parsed back ✅
  - Validation catches missing/invalid fields ✅

No deviations from the plan. All changes are minimal and follow existing patterns in the codebase.
