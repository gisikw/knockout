## Goal
Prevent decision nodes from receiving their own prior output as context, eliminating classification confusion on retry builds.

## Context
The diagnosis already analyzed the issue comprehensively:

- **Root cause**: `build.go:runPromptNode` (lines 466-470) unconditionally injects prior context for ALL prompt nodes
- **Problem**: Decision nodes like `classify` see their own previous output when a ticket is built multiple times, causing Haiku to second-guess classifications
- **Current state**: `InjectPriorContext()` correctly filters by workflow but doesn't distinguish between node types
- **Node types**: `NodeDecision` and `NodeAction` constants already exist in `workflow.go:9-11`
- **Affected workflows**: All decision nodes (`classify`, `actionable`, `assess`, `review`) benefit from NOT receiving prior context; all action nodes should continue receiving it

The pipeline structure from `.ko/pipeline.yml`:
- Main workflow: `classify` (decision) routes to task/research/bug
- Task workflow: `plan` (action) → `actionable` (decision) → `implement` (action) → `verify` (action) → `review` (decision)
- Bug workflow: `diagnose` (action) → `assess` (decision)
- Research workflow: `investigate` (action)

Testing patterns from `workspace_test.go` and `build_test.go` show the project uses table-driven tests with temp directories.

## Approach
Add a conditional check in `runPromptNode` to only inject prior context when `node.Type == NodeAction`. This is a one-line logic change with a comment. Decision nodes make routing/classification choices that should be idempotent; action nodes perform incremental work that benefits from continuity across retries.

## Tasks
1. [build.go:466-470] — Wrap the prior context injection in a conditional `if node.Type == NodeAction`.
   Add a comment explaining that decision nodes should make fresh evaluations without seeing their own previous output.
   Verify: `go build ./...` succeeds.

2. [build_test.go] — Add a test function `TestRunPromptNodePriorContextInjection` that verifies action nodes receive prior context and decision nodes do not.
   Use a mock setup with temp directories containing workspace files.
   Verify: `go test ./... -run TestRunPromptNodePriorContextInjection` passes.

3. [testdata/pipeline/] — Review existing txtar test fixtures to ensure none incorrectly expect prior context in decision node prompts.
   If any do, update them to reflect the new behavior.
   Verify: `go test ./...` passes (all pipeline integration tests).

## Open Questions
None. The fix is straightforward and aligns with the semantic purpose of decision vs action nodes. The diagnosis already confirmed that all current decision nodes would benefit from this change, and no decision nodes need prior context injection.
