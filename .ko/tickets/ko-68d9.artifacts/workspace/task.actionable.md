knockout — just --list for recipes
Looking at the plan document, I need to evaluate it across three dimensions:

**1. Open Questions:**
The plan has one clear open question that requires human input:
- **Migration destructiveness**: Whether to keep the old `projects.yml` file in place (non-destructive, satisfying "no longer read") or delete it (destructive, satisfying "no longer exists").

The second item listed ("Footer ordering") is not actually an open question—it's a directive to ensure task 5 includes test coverage for that edge case.

**2. Completeness:**
The plan is well-structured with:
- Clear goal: consolidate `projects.yml` into `config.yaml`
- Identified files: `registry.go`, `global_config.go`, `cmd_project_test.go`, `registry_test.go`
- 6 ordered tasks with verification steps for each
- Specific functions to modify/add with precise logic descriptions

**3. Feasibility:**
Based on the Context section, the plan is feasible:
- Files and functions mentioned are documented as existing (`RegistryPath()`, `GlobalConfigPath()`, `LoadRegistry()`, `SaveRegistry()`)
- The approach leverages existing tolerant parsers (both handle extra sections already)
- Test paths and patterns are clearly referenced

The plan cannot proceed without resolving the migration destructiveness decision.

```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "Should the old projects.yml be deleted after successful migration, or kept on disk?",
      "context": "Obligation 2 requires projects.yml to 'no longer exist OR is no longer read'. The proposed non-destructive approach keeps the file but stops reading it—satisfying the 'no longer read' clause.",
      "options": [
        {
          "label": "Keep old file (Recommended)",
          "value": "keep_old_file",
          "description": "Non-destructive: projects.yml remains on disk but is never read after migration. Safer for rollback."
        },
        {
          "label": "Delete old file",
          "value": "delete_old_file",
          "description": "Destructive: Explicitly delete projects.yml after successful migration. Fully satisfies obligation, cleaner final state."
        }
      ]
    }
  ]
}
```
