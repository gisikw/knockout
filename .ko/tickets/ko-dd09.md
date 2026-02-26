---
id: ko-dd09
status: resolved
deps: []
created: 2026-02-26T21:43:40Z
type: task
priority: 2
---
# Only inject prior output into action nodes, not decision nodes. Haiku is getting confused classifying tickets that it already classified once

## Notes

**2026-02-26 21:52:48 UTC:** # Implementation Summary: Prior Context Injection for Node Types

## What Was Done

Fixed the issue where decision nodes were receiving their own previous output as "prior context" on retry builds, causing Haiku to second-guess classifications.

### Changes

1. **build.go:466-474** — Added conditional check to only inject prior context for action nodes
   - Wrapped `InjectPriorContext` call in `if node.Type == NodeAction`
   - Added explanatory comments distinguishing decision vs action node semantics
   - Decision nodes now make fresh, idempotent evaluations
   - Action nodes continue receiving prior context for continuity across retries

2. **build_test.go:168-239** — Added `TestRunPromptNodePriorContextInjection`
   - Table-driven test verifying behavior for both `NodeAction` and `NodeDecision`
   - Validates that action nodes receive prior context
   - Validates that decision nodes do NOT receive prior context
   - Uses temp directories and mirrors the exact prompt construction logic from `runPromptNode`

3. **Integration Test Verification** — All 40 pipeline tests pass
   - `build_resume_from_prior_plan.txtar` correctly tests action node prior context injection
   - `build_resume_from_workspace.txtar` correctly tests workspace context for action nodes
   - `build_max_visits.txtar` (has decision node loop) works correctly without prior context
   - No test fixtures required updates

## Notable Decisions

- **Semantic distinction preserved**: Decision nodes make routing/classification choices that should be idempotent and fresh each time. Action nodes perform incremental work that benefits from knowing what was done in prior attempts.

- **Minimal change**: The fix is a simple 4-line conditional (including comments) that leverages existing `NodeType` constants, avoiding any architectural refactoring.

- **Test coverage**: The new unit test directly validates the behavioral contract. Existing integration tests confirm no regressions across all workflows (`main`, `task`, `bug`, `research`).

## Future Considerations

None. The implementation is complete and aligns perfectly with the semantic purpose of the two node types. Decision nodes (`classify`, `actionable`, `assess`, `review`) now make clean evaluations without confusion from their own previous output. Action nodes (`plan`, `implement`, `verify`, `diagnose`, `investigate`) retain continuity across build attempts as intended.

All tests pass. No deviations from plan. Ready to close.

**2026-02-26 21:52:48 UTC:** ko: SUCCEED
