## Goal
Enable agents to resume from existing plan/artifacts on retry instead of starting from scratch.

## Context

**Workspace Structure:**
- Each build creates `.ko/tickets/<id>.artifacts/workspace/`
- Node outputs are tee'd as `<workflow>.<node>.md` (e.g., `task.plan.md`, `task.implement.md`)
- Artifact directory persists across builds (invariant: line 90-95 in INVARIANTS.md)
- Plan files are written to `$KO_ARTIFACT_DIR/plan.md` by plan node
- The implement prompt already instructs: "Read the plan at `$KO_ARTIFACT_DIR/plan.md`. Follow it."

**Current Behavior:**
- `runPromptNode()` (build.go:402-485) constructs prompts with: ticket body, discretion guidance, and prompt content
- No injection of prior workspace outputs happens today
- On retry, agents re-plan from scratch even if `plan.md` exists

**Key Files:**
- `build.go:runPromptNode()` — constructs prompts for agent nodes
- `workspace.go` — workspace creation and output tee logic (28 lines)
- `.ko/prompts/implement.md` — already references `$KO_ARTIFACT_DIR/plan.md`
- `.ko/prompts/plan.md` — instructs to check for prior plan at `$KO_ARTIFACT_DIR/plan.md`

**Testing Patterns:**
- testscript format in `testdata/pipeline/*.txtar`
- Tests use `run:` nodes to verify behavior via shell commands
- `build_artifact_dir.txtar` shows artifact directory persistence across nodes
- `build_retry_exhausted.txtar` shows retry mechanics

## Approach

Add a new function `InjectPriorContext()` that scans for existing workspace task files and plan.md in the artifact directory. Inject this as a "## Prior Context" section between the ticket and instructions in `runPromptNode()`. This gives agents visibility into what was already done without changing prompt files or the workspace persistence model.

The injection is workflow-aware: only inject workspace files from the current workflow (e.g., for `task.implement`, inject `task.plan.md` but not `bug.diagnose.md`).

## Tasks

1. [workspace.go:InjectPriorContext] — Create function that scans artifactDir for `plan.md` and workspace files matching the current workflow prefix (e.g., `task.*`). Returns a formatted markdown string with file contents, or empty string if no prior context found.
   Verify: `go test ./... -count=1` passes.

2. [workspace_test.go] — Add `TestInjectPriorContext()` covering: no prior files (returns empty), plan.md only, workspace task files only, multiple workspace files filtered by workflow, and full scenario with plan + workspace files.
   Verify: `go test ./... -count=1` passes, new test passes.

3. [build.go:runPromptNode] — Call `InjectPriorContext(artifactDir, wfName)` after loading promptContent. If non-empty, insert "## Prior Context" section between ticket and instructions.
   Verify: `go test ./... -count=1` passes.

4. [testdata/pipeline/build_resume_from_prior_plan.txtar] — Test that verifies: first build creates plan.md, second build (triggered manually) sees and references prior plan. Use a run node that checks if plan.md exists and outputs "found prior plan".
   Verify: `ko test` passes, new test passes.

5. [testdata/pipeline/build_resume_from_workspace.txtar] — Test that verifies: first build writes workspace task files, second build on same ticket injects them. Use a run node to verify workspace files are visible in subsequent node execution.
   Verify: `ko test` passes, new test passes.

## Open Questions

**Q: Should we inject ALL workspace files or only those from the current workflow?**
A: Only current workflow. If a ticket was classified, routed to bug, diagnosed, then re-routed to task — the task nodes shouldn't see bug.diagnose output. This prevents context pollution across workflow boundaries.

**Q: What's the injection format?**
A: Use markdown code fences with file labels:
```
## Prior Context

From previous build attempts:

### plan.md
<content>

### task.plan.md
<content>

### task.implement.md
<content>
```

**Q: Should retries within the same build see prior workspace outputs from earlier nodes?**
A: No. Workspace outputs are tee'd after each node completes successfully (build.go:208). Retries happen within `runNode()` before TeeOutput is called. This feature is for build-to-build resume, not intra-build visibility (which already works via $KO_ARTIFACT_DIR).

**Q: File size limits for injection?**
A: Use `os.ReadFile` without limits for now. Workspace files are agent outputs (typically < 10KB). If this becomes a problem in practice, add truncation later with a visible marker.

**Q: Should we update prompt files to reference this feature?**
A: No. The prompts already reference `$KO_ARTIFACT_DIR` for reading artifacts. This injection is an optimization — agents can still read files via env var if needed. Keep prompts stable.
