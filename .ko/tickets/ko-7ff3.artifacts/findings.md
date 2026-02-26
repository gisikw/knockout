# Findings: Go Command Organization

## Current State

The `knockout` project currently has **15 command files** at the top level using the `cmd_*.go` naming pattern:

```
cmd_agent.go         (456 lines)
cmd_build.go         (101 lines)
cmd_build_init.go    (211 lines)
cmd_bump.go          (38 lines)
cmd_create.go        (358 lines)
cmd_dep.go           (244 lines)
cmd_list.go          (285 lines)
cmd_loop.go          (221 lines)
cmd_note.go          (80 lines)
cmd_project.go       (207 lines)
cmd_serve.go         (146 lines)
cmd_serve_sse.go     (591 lines)
cmd_show.go          (208 lines)
cmd_status.go        (104 lines)
cmd_update.go        (245 lines)
```

These live alongside 14 core domain files (ticket.go, build.go, pipeline.go, etc.) in a single flat `main` package with **~6,954 total lines** of non-test code.

## Analysis

### Current Pattern: Flat with `cmd_` Prefix

**Advantages:**
1. **Simple imports** - Everything is in `main` package, no import paths needed
2. **Fast navigation** - `cmd_*.go` pattern is easy to glob/grep
3. **Low ceremony** - No subdirectory structure, no package boilerplate
4. **Clear naming** - The `cmd_` prefix explicitly signals "this is a command handler"
5. **Testable** - Test files are `cmd_*_test.go` right next to implementation
6. **Compliant with INVARIANTS.md** - Splits along behavioral seams (each command is a behavioral unit)

**Disadvantages:**
1. **Visual clutter** - 29 Go files at top level (15 cmd + 14 core)
2. **No hierarchical organization** - All commands appear equal in directory listing
3. **Scaling concern** - Adding more commands continues to expand the flat list

### Alternative: Nested `cmd/` or `internal/cmd/` Package

Standard Go convention for multi-command CLIs is:
```
cmd/
  agent/
    agent.go
  build/
    build.go
  ...
main.go
```

Or for single-binary CLIs like `knockout`:
```
internal/cmd/
  agent.go
  build.go
  ...
main.go
```

**Advantages of nesting:**
1. **Visual separation** - Commands live in their own directory
2. **Scalability** - Can add subpackages if a command grows (e.g., `cmd/agent/` subpackage)
3. **Standard Go pattern** - Matches how most Go CLIs are structured (kubectl, docker, gh, etc.)

**Disadvantages of nesting:**
1. **Import overhead** - Need to import the cmd package from main
2. **Export requirements** - Command functions must be capitalized: `CmdAgent` vs `cmdAgent`
3. **More file tree depth** - One more level to navigate
4. **Split tests** - Test files might end up in cmd/ away from core logic
5. **Not actually needed** - This is a single binary, not multiple commands needing isolation

## Key Project Constraints

From **INVARIANTS.md Line 122-138**:
- **500 lines max per file** - Current files comply (largest is 591 lines, slightly over)
- **Split along behavioral seams** - Each command is a behavioral unit
- **Tests mirror source files** - `cmd_*.go` → `cmd_*_test.go` works well
- **No multi-purpose functions** - Commands are thin orchestrators over domain logic

From **INVARIANTS.md Line 39-41**:
- **Zero external runtime dependencies** - Single static binary
- No multi-binary setup needed

## Comparable Projects

Standard Go CLI patterns:
- **kubectl** - Uses `cmd/` with subpackages for each command
- **docker** - Uses `cmd/docker/` with subpackages
- **gh (GitHub CLI)** - Uses flat `pkg/cmd/` with one file per command group
- **hugo** - Uses flat `commands/` package with `cmd_*.go` style files
- **rclone** - Uses flat `cmd/` with one file per command

**Pattern observation**: Large CLIs (20+ commands) tend to use `cmd/` subdirectories. Small-to-medium CLIs (10-20 commands) often stay flat or use a single `cmd` package.

## Technical Debt Check

Current violations:
- `cmd_serve_sse.go` is **591 lines** (exceeds 500-line limit by 91 lines)
- Otherwise, all command files comply with size limits

## Recommendation

**KEEP the current flat structure with `cmd_*.go` naming.**

### Rationale:

1. **Simplicity wins** - This is a single-binary CLI tool, not a multi-binary project. The flat structure has lower cognitive overhead.

2. **File size is the real constraint** - The INVARIANTS specify a 500-line limit per file. The current organization respects this except for one outlier. The right fix is to split `cmd_serve_sse.go`, not to reorganize everything.

3. **Behavioral cohesion** - Commands are already split along behavioral seams. Moving them to `cmd/` or `internal/cmd/` doesn't improve cohesion—it just adds a directory level.

4. **Scale is manageable** - 15 commands is not excessive for a flat structure. Even if it grows to 25-30 commands, a flat list with clear naming (`cmd_*.go`) remains navigable.

5. **Test proximity** - Current pattern keeps tests (`cmd_*_test.go`) directly adjacent to implementation. Moving commands to a subdirectory could separate tests from core domain logic.

6. **No multi-package benefit** - The commands don't need package-level isolation from each other. They're all thin orchestrators over the same domain types (Ticket, Pipeline, Registry).

7. **Editor/tooling support** - Modern editors handle flat file lists well with fuzzy finding. The `cmd_` prefix makes glob patterns trivial: `cmd_*.go`, `grep -l "cmdAgent"`, etc.

### Recommended Actions

1. **Split `cmd_serve_sse.go`** (591 lines) to comply with the 500-line limit
   - Extract SSE-specific logic into `sse.go` or similar
   - Keep command handler in `cmd_serve_sse.go`

2. **Document the pattern** - Add to INVARIANTS.md:
   ```markdown
   ## CLI Commands
   - **Commands use `cmd_*.go` naming at package root.** Each user-facing
     command has a `cmdFoo(args []string) int` function in a `cmd_foo.go` file.
   - **Keep commands flat.** No `cmd/` subdirectory unless we exceed 30+ commands.
     Flat structure reduces import ceremony and keeps tests adjacent to domain logic.
   - **Commands are thin orchestrators.** Command functions handle flag parsing
     and orchestration. Business logic lives in domain files (ticket.go, build.go, etc.).
   ```

3. **Consider command grouping** - If the list continues to grow, consider functional grouping:
   - `cmd_ticket_*.go` for ticket operations (create, show, list, status)
   - `cmd_agent_*.go` for agent operations (already have cmd_agent.go, cmd_loop.go)
   - `cmd_dep_*.go` for dependency management (already have cmd_dep.go)
   - But this is a future concern, not urgent now

## Alternative Considered: `internal/cmd/` Package

If the command count doubles (30+ commands), consider:
```
internal/cmd/
  agent.go
  build.go
  create.go
  ...
main.go
```

This would require:
- Exporting command functions (`CmdAgent` instead of `cmdAgent`)
- Importing from main: `import "git.gisi.network/infra/knockout/internal/cmd"`
- Possibly exporting domain types or creating an internal API package

**Cost/benefit**: Adds complexity without clear benefit at current scale. Revisit if command count exceeds 25-30.

## Summary

**Answer: No, nesting is not necessary.**

The flat structure with `cmd_*.go` naming is appropriate for this project's size and constraints. The real organization issue is file size, not directory structure. Fix the one file that exceeds the line limit, document the pattern in INVARIANTS.md, and continue with the current approach.

The flat pattern trades some visual hierarchy for simplicity and low ceremony, which aligns well with the project's single-binary, zero-dependency philosophy. If the command surface area grows significantly (2x or more), reassess—but for now, the current structure is sound.
