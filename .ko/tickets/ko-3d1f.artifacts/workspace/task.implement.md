Excellent! All tests pass. Let me create a brief summary of what was implemented:

## Implementation Summary

I successfully implemented the `ko questions` command following the approved plan. Here's what was changed:

### Task 1: Created cmd_questions.go
- **File**: `cmd_questions.go`
- Created new command function `cmdQuestions()` that:
  - Accepts a ticket ID as argument
  - Uses `resolveProjectTicketsDir()` for project resolution
  - Uses `ResolveID()` for partial ID matching
  - Loads the ticket and marshals `PlanQuestions` to JSON with indentation
  - Outputs to stdout
  - Returns appropriate exit codes (0 for success, 1 for errors)

### Task 2: Added 'questions' case to main.go
- **File**: `main.go:62-63`
- Added `case "questions":` to the switch statement in the `run()` function
- Placed alphabetically between "bump" and "query" commands
- Calls `cmdQuestions(rest)` and returns the result

### Task 3: Added help text
- **File**: `main.go:148`
- Added `questions <id>     Show plan questions as JSON` to the help text
- Positioned after the `block` command (as questions are related to blocking)

### Task 4: Created comprehensive unit tests
- **File**: `cmd_questions_test.go`
- Implemented three test functions:
  - `TestCmdQuestions`: Tests tickets with plan-questions, no plan-questions, and nil plan-questions
  - `TestCmdQuestionsErrors`: Tests error cases (missing ID, nonexistent ticket)
  - `TestCmdQuestionsPartialID`: Tests partial ID resolution
- All tests pass successfully

### Verification
- ✅ Code compiles without errors
- ✅ `ko questions` shows usage error when called without arguments
- ✅ `ko help` displays the new command
- ✅ All unit tests pass (3 test functions, 6 test cases total)
- ✅ All existing tests still pass

The implementation follows the existing codebase patterns, complies with INVARIANTS.md, and provides the JSON output format required for the Punchlist server to render question UI for blocked tickets.
