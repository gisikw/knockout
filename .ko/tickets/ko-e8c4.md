---
id: ko-e8c4
status: blocked
deps: [ko-1bfc, ko-b99d, ko-dbe4, ko-70b8, ko-5734, ko-7537, ko-860e, ko-88fa]
created: 2026-02-25T20:11:38Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Should we add a migration guide or deprecation note to the README explaining the removed commands?"
    context: "The CLI consolidation removed several commands (create, closed, resolved, add-note, reopen, query, questions, answer, block, register, init, default, projects). Users with scripts using the old syntax will get 'unknown command' errors."
    options:
      - label: "Add migration guide section (Recommended)"
        value: add_migration_guide
        description: "Include a brief 'Migration from older CLI versions' section explaining what changed and how to update old scripts"
      - label: "No migration guide"
        value: no_migration_guide
        description: "Only update command listings; users can infer changes from the new command surface and error messages"
---
# Update README and help text for CLI simplification

After all CLI consolidation tickets are done, update README.md and ko help output to reflect the simplified command surface.
