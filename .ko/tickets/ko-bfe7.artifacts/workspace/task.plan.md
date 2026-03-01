The plan is written to `/home/dev/Projects/knockout/.ko/tickets/ko-bfe7.artifacts/plan.md`. Here's a summary:

**Goal**: Replace the flat three-section `projects.yml` (`default:` top-level + `projects:` + `prefixes:`) with a single nested structure where each project's `path`, `prefix`, and `default` are all grouped under the project tag.

**New format:**
```yaml
projects:
  exo:
    path: /home/dev/Projects/exocortex
    prefix: exo
    default: true
  knockout:
    path: /home/dev/Projects/knockout
    prefix: ko
```

**Key decisions:**
- The `Registry` Go struct is unchanged — only serialization changes, so all 5+ callers (`cmd_project.go`, `cmd_list.go`, `ticket.go`, etc.) need no changes
- Old format remains parseable for backward compatibility
- `LoadRegistry` auto-migrates old-format files to new format on first read
- All inline test YAML strings in `registry_test.go` and `cmd_list_test.go` get updated to new format

**6 tasks**, all in `registry.go` and test files:
1. Update `ParseRegistry` — handle nested format while keeping old format working
2. Update `FormatRegistry` — emit new nested format
3. Add `isOldFormat` + auto-migration in `LoadRegistry`
4. Update `registry_test.go` inline YAML
5. Update `cmd_list_test.go` inline YAML
6. Full test suite verification
