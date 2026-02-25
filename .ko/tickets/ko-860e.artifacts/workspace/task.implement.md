Excellent! All tasks have been completed successfully. Let me provide a summary of the changes made:

## Implementation Summary

I successfully implemented the ticket according to the plan, consolidating `ko block`, `ko questions`, and `ko answer` into a single `ko triage` command. Here's what was changed, organized by task:

### Task 1: Created cmd_triage.go
- Implemented `cmdTriage` function with four modes:
  - Bare invocation: shows block reason (via `ExtractBlockReason`) and questions as JSON
  - `--block [reason]`: blocks ticket with optional parseable note format "ko: BLOCKED — {reason}"
  - `--questions '<json>'`: validates and adds questions, implicitly sets status to blocked
  - `--answers '<json>'`: answers questions with notes, auto-unblocks when all answered
- Reused existing validation and logic patterns from the old commands

### Task 2: Added triage case to main.go
- Added `case "triage":` to the command switch statement

### Task 3: Created cmd_triage_test.go
- Comprehensive test coverage including:
  - Bare invocation tests (with/without block reasons and questions)
  - Block tests (with/without reasons)
  - Questions tests (valid, invalid JSON, validation errors)
  - Answers tests (partial, full, error cases)
  - Error handling tests
- All tests pass

### Tasks 4-9: Removed old commands
- Removed `block`, `questions`, and `answer` cases from main.go switch statement
- Deleted `cmdBlock` function from cmd_status.go (kept `ValidatePlanQuestions` for reuse)
- Deleted files: cmd_questions.go, cmd_questions_test.go, cmd_answer.go, cmd_answer_test.go
- Updated testdata files to use new `ko triage` command syntax:
  - testdata/ticket_status/status_ready_exclusion.txtar
  - testdata/ticket_status/status_shortcuts.txtar
  - testdata/plan_questions/answer_errors.txtar
  - testdata/plan_questions/answer_full.txtar
  - testdata/plan_questions/answer_partial.txtar

### Task 10: Updated help text in main.go
- Removed help entries for `block`, `questions`, and `answer` commands
- Added four new help entries documenting all `ko triage` usage modes

### Task 11: Updated README.md
- Updated "Plan Questions" section with new `ko triage` command examples
- Added example for blocking with reason

### Verification
- All tests pass (`go test ./...`)
- Build succeeds without errors
- Help text correctly displays new command
- No references to old commands remain in code or documentation

The implementation follows the revised design from the ticket notes, using implicit status transitions (--questions → blocked, --answers resolving all questions → open) and the parseable "ko: BLOCKED — {reason}" note format for ExtractBlockReason compatibility.
