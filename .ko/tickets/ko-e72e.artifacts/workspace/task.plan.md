Excellent! The plan has been successfully written to `$KO_ARTIFACT_DIR/plan.md`. 

## Summary

I've created a comprehensive implementation plan for enabling agents to resume from existing plan/artifacts on retry. The plan includes:

### Goal
Enable agents to reuse existing plans and workspace artifacts from prior build attempts instead of starting from scratch on retries.

### Key Approach
- Add an `InjectPriorContext()` function that scans for existing workspace files and plan.md
- Inject this as a "Prior Context" section in the prompt between ticket and instructions
- Make it workflow-aware to prevent context pollution across workflow boundaries

### Implementation Tasks
1. Create `InjectPriorContext()` function in workspace.go
2. Add comprehensive tests in workspace_test.go
3. Integrate the injection into `runPromptNode()` in build.go
4. Add two testscript tests to verify the resume behavior

### Design Decisions (resolved in Open Questions)
- Only inject files from the current workflow to avoid context pollution
- Use markdown code fences with clear file labels
- No file size limits initially (workspace files are typically small)
- No prompt file updates needed (they already reference `$KO_ARTIFACT_DIR`)

The plan is proportional to the ticket size, provides concrete verification steps for each task, and maintains the architectural separation between decision logic and I/O that's documented in INVARIANTS.md.
