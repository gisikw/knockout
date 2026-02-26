---
id: ko-88dd
status: blocked
deps: []
created: 2026-02-26T20:30:31Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Should require_clean_tree: true prevent builds when there ARE uncommitted changes, or when there are NO uncommitted changes?"
    context: "The ticket title says 'prevent if clean' but the flag name and Git conventions suggest 'prevent if dirty'. This determines the core behavior: require_clean_tree: true should either (A) block builds with uncommitted changes, or (B) block builds when everything is committed."
    options:
      - label: "Prevent if dirty (Recommended)"
        value: prevent_if_dirty
        description: "require_clean_tree: true blocks builds when there ARE uncommitted changes (standard interpretation)"
      - label: "Prevent if clean"
        value: prevent_if_clean
        description: "require_clean_tree: true blocks builds when there are NO uncommitted changes (literal ticket title interpretation)"
---
# Add support for require_clean_tree: true in pipeline.yml, which would prevent an agent from starting if the working tree is clean (ignoring ./.ko)
