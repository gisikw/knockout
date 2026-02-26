The plan has a clear **Open Questions** section that requires human clarification before implementation can proceed. The ticket title contains a potential semantic inconsistency that the plan author correctly flagged:

- Ticket title: "prevent an agent from starting if the working tree is **clean**"
- Plan interpretation: `require_clean_tree: true` means prevent builds when tree is **dirty** (has uncommitted changes)

This is a fundamental behavioral difference that must be clarified.

```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "Should require_clean_tree: true prevent builds when there ARE uncommitted changes, or when there are NO uncommitted changes?",
      "context": "The ticket title says 'prevent if clean' but the flag name and Git conventions suggest 'prevent if dirty'. This determines the core behavior: require_clean_tree: true should either (A) block builds with uncommitted changes, or (B) block builds when everything is committed.",
      "options": [
        {
          "label": "Prevent if dirty (Recommended)",
          "value": "prevent_if_dirty",
          "description": "require_clean_tree: true blocks builds when there ARE uncommitted changes (standard interpretation)"
        },
        {
          "label": "Prevent if clean",
          "value": "prevent_if_clean",
          "description": "require_clean_tree: true blocks builds when there are NO uncommitted changes (literal ticket title interpretation)"
        }
      ]
    }
  ]
}
```
