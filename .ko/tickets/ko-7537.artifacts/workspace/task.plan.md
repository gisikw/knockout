Perfect! I've created the implementation plan. Let me summarize what I found and the approach:

## Summary

I've investigated the codebase and created a detailed implementation plan at `.ko/tickets/ko-7537.artifacts/plan.md`.

**Key findings:**
- Four commands currently exist: `init` (cmd_init.go), `register`, `default`, `projects` (all in cmd_registry.go)
- Registry infrastructure is well-established with YAML persistence at `~/.config/knockout/projects.yml`
- Init writes prefix to `.ko/config.yaml` using `WriteConfigPrefix`
- Tests use `t.TempDir()` for isolation and mirror source files

**Implementation approach:**
Create `cmd_project.go` with a dispatcher handling two subcommands:
- `project set #tag [--prefix=p] [--default]` — Upsert operation combining init+register+default behaviors
- `project ls` — Unified listing showing all projects with default marker (asterisk)

The plan has 8 tasks covering implementation, testing, dispatcher updates, help text changes, cleanup of old commands, and integration verification. All existing registry helper functions remain unchanged.
