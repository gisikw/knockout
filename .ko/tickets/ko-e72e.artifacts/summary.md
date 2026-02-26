# Implementation Summary: Agent Resume from Prior Context

## What Was Done

Implemented automatic injection of prior build artifacts into agent prompts on retry. When a ticket is retried after failure, agents now see the plan and workspace outputs from previous attempts, eliminating redundant planning work.

### Core Changes

1. **workspace.go** - Added `InjectPriorContext()` function (52 lines)
   - Scans artifact directory for `plan.md` and workflow-specific workspace files
   - Returns formatted markdown with "## Prior Context" header
   - Filters workspace files by workflow prefix to prevent cross-workflow context pollution
   - Skips empty files to avoid noise

2. **build.go** - Modified `runPromptNode()` to inject prior context
   - Calls `InjectPriorContext()` after loading prompt content
   - Injects between ticket/discretion and instructions sections
   - Only 6 lines added to existing function

3. **workspace_test.go** - Comprehensive unit tests (134 lines)
   - Covers all scenarios: no prior files, plan only, workspace only, filtering, full scenario
   - Includes edge case testing for empty files
   - All tests pass

4. **Integration tests** - Two testscript tests added
   - `build_resume_from_prior_plan.txtar` - Verifies plan.md injection across builds
   - `build_resume_from_workspace.txtar` - Verifies workspace file injection and filtering
   - Both tests pass

## Key Decisions

**Workflow-aware filtering:** Only workspace files matching the current workflow prefix are injected (e.g., `task.*` files for task workflow). This prevents context pollution if a ticket was routed between workflows.

**Injection placement:** Prior context is inserted between discretion guidance and instructions. This gives agents visibility into previous work while keeping the prompt structure stable.

**No file size limits:** Used `os.ReadFile` without truncation. Workspace files are typically small (<10KB). If this becomes an issue, truncation can be added later.

**No prompt file updates:** Existing prompts already reference `$KO_ARTIFACT_DIR` for reading artifacts. This injection is an optimization—agents can still read files manually if needed.

## Compliance Notes

- All planned tasks completed ✓
- All tests pass (unit and integration) ✓
- No invariant violations introduced ✓
- Pre-existing issue: `build.go` is 765 lines (exceeds 500-line invariant from line 123-138), but this was not introduced by this change (only +6 lines)
- Implementation matches plan exactly—no deviations ✓

## Future Considerations

- If workspace files grow large, consider adding truncation with visible markers
- Could extend to inject history.jsonl excerpts for retry context
- Artifact directory cleanup logic (lines 90-95 invariant) remains unchanged
