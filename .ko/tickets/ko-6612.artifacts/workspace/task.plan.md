Perfect! I've written a comprehensive plan for implementing the `ko update` command. The plan is now saved at `.ko/tickets/ko-6612.artifacts/plan.md`.

## Summary

I investigated the codebase and created an implementation plan with the following key findings:

**Context discovered:**
- Clear command pattern: `cmd_*.go` files with matching tests
- All commands use `resolveProjectTicketsDir()` and `ResolveID()` utilities
- Auto-unblocking logic already exists in `cmd_triage.go` for answering questions
- Tag replacement (not append) is explicitly required per the ticket spec

**Implementation approach:**
- Create `cmd_update.go` following established patterns from `cmd_create.go` and `cmd_triage.go`
- Reuse existing validation logic (`ValidatePlanQuestions()`) 
- Implement auto-unblocking: when `--answers` resolves all questions on a blocked ticket, transition status to `open`
- Wire into main dispatcher and help text

**4 tasks identified:**
1. Create `cmd_update.go` with full flag parsing and field mutation logic
2. Add command dispatcher case in `main.go`
3. Add help text in `cmdHelp()`
4. Create `cmd_update_test.go` with coverage for field updates and auto-unblock behavior

The plan is ready for implementation with no open questions.
