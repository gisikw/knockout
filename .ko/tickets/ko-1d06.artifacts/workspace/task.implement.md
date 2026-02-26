Perfect! All tests pass. Let me create a quick summary of the implementation.

---

## Summary

I successfully implemented all 8 tasks from the plan to add `on_loop_complete` hook support to the knockout pipeline system. Here's what was changed:

### Task 1: Pipeline struct (pipeline.go:35-38)
- Added `OnLoopComplete []string` field to the Pipeline struct after OnClose
- Added comment explaining these hooks run after the loop completes

### Task 2: YAML parsing (pipeline.go:142, 190-193, 361-366)
- Updated section comment to include "on_loop_complete" 
- Added case for "on_loop_complete:" section header
- Added parsing logic in the switch statement to append commands to `p.OnLoopComplete`

### Task 3: Loop hooks function (loop.go:3-8, 142-188)
- Added required imports (os/exec, strconv)
- Created `runLoopHooks` helper function that:
  - Accepts ticketsDir, hooks, LoopResult, and elapsed duration
  - Sets environment variables: LOOP_PROCESSED, LOOP_SUCCEEDED, LOOP_FAILED, LOOP_BLOCKED, LOOP_DECOMPOSED, LOOP_STOPPED, LOOP_RUNTIME_SECONDS
  - Expands variables in hook commands
  - Executes hooks with proper environment

### Task 4: Hook invocation (cmd_loop.go:218-222)
- Called `runLoopHooks` after `writeAgentLogSummary` (line 215) but before cleanup
- Logs errors to stderr but doesn't change exit code (as per plan)

### Task 5: Specification (specs/loop.feature:88-105)
- Added two new scenarios:
  - "on_loop_complete hooks run after loop finishes" - validates env vars are set correctly
  - "on_loop_complete hooks run regardless of stop reason" - validates hooks run for different stop reasons

### Task 6: Test (testdata/loop/loop_on_complete_hook.txtar)
- Created txtar test with pipeline config containing on_loop_complete hooks
- Hooks write all env vars to a file
- Test verifies hook runs and all env vars are set correctly
- Test passes: ✅

### Task 7: Example (examples/default/pipeline.yml:40-46)
- Added commented-out example with clear explanation
- Shows available env vars
- Demonstrates conditional logic (only push if succeeded > 0)

### Task 8: Documentation (README.md:315-319)
- Added `on_loop_complete` to Hooks section
- Documented when hooks run (after loop completes, regardless of stop reason)
- Listed all available env vars
- Noted that hook failures don't affect loop exit code

### Verification
- All tasks completed in order
- All verification steps passed
- Full test suite passes (go test ./...)
- No deviations from the plan

The implementation follows existing patterns for hooks (similar to `runHooks` in build.go) and maintains consistency with the codebase architecture per INVARIANTS.md.
