# Implementation Summary: ko answer command

## What Was Done

Successfully implemented the `ko answer` command for submitting plan question answers and unblocking tickets.

### Files Created
- **cmd_answer.go** - Main command implementation (105 lines)
- **cmd_answer_test.go** - Table-driven Go tests (232 lines)
- **testdata/plan_questions/answer_partial.txtar** - Testscript test for partial answers
- **testdata/plan_questions/answer_full.txtar** - Testscript test for full answer unblocking
- **testdata/plan_questions/answer_errors.txtar** - Testscript test for error cases

### Files Modified
- **main.go** - Added `"answer"` case to command router
- **specs/ticket_status.feature** - Added 5 behavioral scenarios for the answer command

## Implementation Details

The command follows the established pattern in the codebase:
1. Parses args to extract ticket ID and JSON payload
2. Validates JSON structure and question IDs
3. Appends timestamped notes for each answered question: `Plan answer (qID): Question → Answer`
4. Removes answered questions from `ticket.PlanQuestions` slice
5. If all questions answered, clears `plan-questions` field and sets status to `open`
6. If questions remain, status stays `blocked`
7. Saves ticket and emits mutation event with type "answer"

### Test Coverage
- ✅ Partial answers (some questions remain, status stays blocked)
- ✅ Full answers (all questions answered, status set to open, plan-questions cleared)
- ✅ Invalid JSON handling
- ✅ Nonexistent question ID validation
- ✅ Ticket with no plan-questions error case

## Compliance Actions Taken

**INVARIANTS.md Compliance:** The initial implementation included Go tests but lacked the required behavioral specs and testscript tests per INVARIANTS.md lines 16-31 ("Every behavior has a spec" and "Every spec has a test").

Added:
- 5 behavioral scenarios to `specs/ticket_status.feature` covering all major use cases
- 3 testscript test files in `testdata/plan_questions/` validating the behavioral specs

Note: The testscript test infrastructure appears to not be fully wired up yet (no TestScript test function found), but the `.txtar` files are in place for when the infrastructure is enabled.

## Notable Decisions

1. **JSON format**: Uses simple `map[string]string` for question ID to answer mapping, no nested structures
2. **Note format**: Consistent with existing note patterns, includes question ID in parentheses for traceability
3. **Status transition**: Only transitions to `open` when ALL questions are answered, otherwise stays `blocked`
4. **Error handling**: Validates all question IDs before making any changes, ensuring atomic operation
5. **Event emission**: Uses "answer" as the mutation event type for downstream consumption

## Future Reader Notes

- The command supports partial answers by design - users can answer questions incrementally
- Question IDs are embedded in note text for audit trail and debugging
- The mutation event can be used by build pipeline or other tooling to react to answered questions
- Answer text is free-form (not validated against option values), allowing users to provide custom responses beyond the predefined options
