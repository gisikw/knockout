# Investigation: Ticket Title in Commit Messages

## Summary

**Yes, we can leverage ticket titles in commit messages.** The ticket title is available but not currently exposed as an environment variable. Adding `TICKET_TITLE` support requires minimal changes to the `runHooks` function.

## Current Environment Variables

The pipeline currently exposes these environment variables to `on_succeed`, `on_fail`, and `on_close` hooks:

| Variable | Description | Source |
|----------|-------------|--------|
| `TICKET_ID` | Ticket identifier (e.g., "ko-4b72") | `build.go:682` |
| `CHANGED_FILES` | Space-separated list of modified files | `build.go:683` |
| `KO_TICKET_WORKSPACE` | Build workspace directory path | `build.go:684` |
| `KO_ARTIFACT_DIR` | Ticket artifacts directory path | `build.go:685` |
| `KO_BUILD_HISTORY` | Build history log path | `build.go:686` |

These are set in `build.go:660-694` within the `runHooks` function.

## Current Commit Message Format

Current `on_succeed` hook (`.ko/pipeline.yml:66-67`):
```yaml
on_succeed:
  - git add -A && git commit -m "ko: ${TICKET_ID}"
```

This produces commits like:
- `ko: ko-0c04`
- `ko: ko-a6b2`

These are functional but not descriptive.

## Ticket Title Availability

The `Title` field IS available in the `*Ticket` struct passed to `runHooks`:

**Location**: `ticket.go:24-43`
```go
type Ticket struct {
    ID            string         `yaml:"id"`
    Status        string         `yaml:"status"`
    // ... other fields ...

    // Title is extracted from the first markdown heading.
    Title string `yaml:"-"`
    // Body is everything after the frontmatter and title.
    Body string `yaml:"-"`
}
```

**Title Extraction**: `ticket.go:305-314`
- Parsed from first `# ` heading in ticket markdown
- Example: `# Ensure ko show [id] --json includes questions json`
- Already loaded before `runHooks` is called

## Implementation Requirements

To expose `TICKET_TITLE` as an environment variable:

### 1. Add to `os.Expand` switch (build.go:668-677)
```go
expanded := os.Expand(hook, func(key string) string {
    switch key {
    case "TICKET_ID":
        return t.ID
    case "TICKET_TITLE":
        return t.Title  // ADD THIS
    case "CHANGED_FILES":
        return changedFiles
    default:
        return os.Getenv(key)
    }
})
```

### 2. Add to cmd.Env (build.go:681-687)
```go
cmd.Env = append(os.Environ(),
    "TICKET_ID="+t.ID,
    "TICKET_TITLE="+t.Title,  // ADD THIS
    "CHANGED_FILES="+changedFiles,
    "KO_TICKET_WORKSPACE="+wsDir,
    "KO_ARTIFACT_DIR="+ArtifactDir(ticketsDir, t.ID),
    "KO_BUILD_HISTORY="+histPath,
)
```

### 3. Update documentation
- `README.md:308` — Add `TICKET_TITLE` to env var list
- `cmd_build_init.go:120` — Update comment showing available env vars
- `examples/default/pipeline.yml:31-34` — Update example comment

## Example Usage

With `TICKET_TITLE` available, the pipeline could use:

```yaml
on_succeed:
  - git add -A && git commit -m "${TICKET_TITLE}" -m "ko: ${TICKET_ID}"
```

This would produce commits like:
```
Ensure ko show [id] --json includes questions json

ko: ko-0c04
```

Or a more concise single-line format:
```yaml
on_succeed:
  - git add -A && git commit -m "ko: ${TICKET_TITLE} (${TICKET_ID})"
```

Producing:
```
ko: Ensure ko show [id] --json includes questions json (ko-0c04)
```

## Edge Cases

1. **Empty titles**: Some tickets might lack a `# ` heading
   - Current behavior: `t.Title` would be empty string
   - Could add fallback: `if t.Title == "" { return t.ID }`

2. **Long titles**: Some titles may exceed conventional commit message length
   - Git convention: ~50 chars for subject line
   - Could truncate: `if len(t.Title) > 50 { return t.Title[:47] + "..." }`
   - Or use full title in body instead of subject

3. **Special characters**: Titles may contain quotes, newlines, or shell metacharacters
   - Need shell escaping for variable expansion
   - Using env vars (vs command-line args) already provides some safety
   - May need additional escaping for git commit -m

## Recommended Actions

1. **Add `TICKET_TITLE` env var** to `build.go:runHooks` (3 line changes)
2. **Update documentation** to list `TICKET_TITLE` in available env vars (3 files)
3. **Update `.ko/pipeline.yml`** to use descriptive commit messages
4. **Add tests** for title handling edge cases (empty, long, special chars)

## References

- Hook execution: `build.go:660-694`
- Ticket struct: `ticket.go:24-43`
- Title parsing: `ticket.go:305-314`
- Current pipeline: `.ko/pipeline.yml:66-67`
- Documentation: `README.md:308`
