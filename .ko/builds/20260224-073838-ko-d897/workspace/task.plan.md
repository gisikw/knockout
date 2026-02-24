I've created a plan for implementing the ticket description feature. The plan includes:

**Goal**: Enable `ko add` to accept descriptions via stdin, heredoc, or second positional arg.

**Key findings**:
- The dependency (ko-90ba) is already closed, so we can implement against `ko add`
- There's a reference implementation in `cmd_note.go` that handles stdin/args similarly
- The current `-d` flag already appends to ticket body, suggesting descriptions should merge

**Approach**: Follow the same stdin detection pattern used in `cmdAddNote()`, with description sources merging into the ticket body.

**Tasks**:
1. Add stdin detection to `cmdCreate()` after title extraction
2. Merge description sources (stdin + second arg + `-d` flag)
3. Add test coverage for all input methods
4. Update or create spec file for documentation

**Open Question**: I've identified one assumption about merging multiple description sources. The current `-d` flag appends to body, so I'm assuming all sources (stdin, second arg, `-d`) should merge together. If strict precedence is preferred instead, that should be clarified.
