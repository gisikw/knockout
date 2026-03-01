---
id: ko-bfe7
status: blocked
deps: []
created: 2026-03-01T06:17:26Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Should knockout auto-migrate users' existing projects.yml files from the old format to new format on first read, or should users manually migrate their config?"
    context: "Task 3 adds auto-migration logic to LoadRegistry that silently rewrites old-format files to the new nested format on first invocation. The plan recommends this for simplicity, but it's an optional feature that could be skipped."
    options:
      - label: "Silent auto-migration (Recommended)"
        value: auto_migrate
        description: "Implement task 3: auto-detect old format and rewrite config file automatically on first read"
      - label: "Manual migration only"
        value: manual_migration
        description: "Skip task 3: users must manually update their config; LoadRegistry only handles reading both formats"
---
# Restructure the projects.yml to avoid redundant lists of things. Everything (path, default, prefix) can just be aggregated under a project name. Update everywhere in knockout that has dependencies on this file structure
