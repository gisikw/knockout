Excellent! All pipeline tests pass, including:
- `build_resume_from_prior_plan` - which tests that action nodes receive prior context ✓
- `build_resume_from_workspace` - which tests that action nodes receive workspace context ✓
- `build_max_visits` - which has a decision node that loops (now won't see its own output) ✓
- All other decision node tests pass ✓

All three tasks are complete! Let me create a summary:

---

## Implementation Summary

Successfully implemented the fix to prevent decision nodes from receiving their own prior output as context.

### Task 1: Modified build.go (lines 466-474)
**File**: `build.go:466-474`

Added a conditional check to only inject prior context for action nodes:
- Wrapped the `InjectPriorContext` call in `if node.Type == NodeAction`
- Added explanatory comments about why decision nodes should make fresh evaluations
- Verification: `go build ./...` succeeded ✓

### Task 2: Added test in build_test.go
**File**: `build_test.go:169-226`

Created `TestRunPromptNodePriorContextInjection` with table-driven tests that verify:
- Action nodes DO receive prior context from `InjectPriorContext`
- Decision nodes do NOT receive prior context
- Used the same prompt construction logic as `runPromptNode` (lines 454-477)
- Verification: `go test ./... -run TestRunPromptNodePriorContextInjection` passed ✓

### Task 3: Reviewed testdata/pipeline/ fixtures
**Reviewed**: 40 txtar test files in `testdata/pipeline/`

Key findings:
- Only 2 tests explicitly check for "Prior Context" injection:
  - `build_resume_from_prior_plan.txtar` - uses action node (`implement`) ✓
  - `build_resume_from_workspace.txtar` - uses action nodes (`plan`, `implement`) ✓
- Decision node tests (e.g., `build_max_visits.txtar`) don't rely on prior context being present
- No test fixtures needed updating
- Verification: `go test -run TestPipeline` passed (all 40 pipeline tests) ✓

### No Deviations
The implementation follows the plan exactly. The fix is minimal, focused, and aligns with the semantic distinction between decision nodes (idempotent routing) and action nodes (incremental work with continuity).
