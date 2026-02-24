Excellent! All tasks completed successfully. Let me provide a summary of the changes:

## Summary

Successfully implemented skill invocation support for pipeline workflows according to the plan. All tasks completed:

### Task 1: Node struct fields (workflow.go:14-24)
- Added `Skills []string` field to store skill directories
- Added `Skill string` field to specify a skill to invoke
- Both marked with comments noting future multi-agent harness support

### Task 2: Validation logic (workflow.go:63-82)
- Updated validation to ensure `Skill` is mutually exclusive with `Prompt` and `Run`
- Modified error messages to reflect all three options (prompt, run, or skill)
- Skills list can coexist with prompt nodes

### Task 3: Parser updates (pipeline.go)
- Added `inSkills` boolean flag (line 76) parallel to `inRoutes`
- Added skill entry handling for multi-line lists (lines 191-194)
- Updated `applyNodeProperty` signature to accept both `inRoutes` and `inSkills` pointers
- Added "skills" case for list parsing (inline `[a, b]` or multi-line)
- Added "skill" case for scalar parsing

### Task 4: Tests (pipeline_test.go:387-491)
- `TestParsePipelineSkillsMultiline`: Tests multi-line skills list and skill field
- `TestParsePipelineSkillsInline`: Tests inline `[a, b]` list syntax
- `TestValidateWorkflowsSkillExclusivity`: Tests validation errors for skill+prompt and skill+run
- Updated workflow_test.go error message to match new validation text

### Task 5: Runtime error (build.go:348-351)
- Added check at start of `runPromptNode` to detect skill invocation
- Returns clear error: "skill invocation not yet supported by this adapter (pending ko-1930)"
- Includes comment documenting the limitation

### Task 6: Documentation (.ko/pipeline.yml:10-19)
- Added commented example section showing all three usage patterns:
  1. `skills:` list to make skills available
  2. `skill:` field to invoke a specific skill
  3. Inline prompt references to skills

All tests pass and the build succeeds. No deviations from the plan.
