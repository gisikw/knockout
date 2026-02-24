Plan written. I've investigated the codebase and created a detailed implementation plan that:

**Key findings:**
- Agent adapters are hardcoded in `adapter.go` with three implementations (ClaudeAdapter, CursorAdapter, RawCommandAdapter)
- The pipeline config currently supports `agent: name` or `command: raw` fields
- Invocation happens in `build.go:runPromptNode()` which calls `adapter.BuildCommand()` and sets env vars

**Proposed approach:**
Extract the hardcoded adapters into YAML config files with a template-based approach. Built-in harnesses (claude, cursor) will be embedded in the binary. Users can add custom harnesses in `~/.config/knockout/agent-harnesses/` or project-local `.ko/agent-harnesses/`.

**Implementation tasks:**
1. Create YAML harness configs for claude and cursor
2. Implement harness loader and template-based adapter
3. Replace hardcoded LookupAdapter switch with harness loading
4. Add tests for harness loading and rendering
5. Embed built-in harnesses in the binary
6. Update documentation

**Open questions** cover:
- YAML schema for conditional args (recommend empty string omission)
- stdin vs arg-based prompt passing (recommend special ${prompt} variable)
- Project-local harness location (recommend `.ko/agent-harnesses/`)
- Harness lookup precedence (recommend project > user > built-ins)

All questions have recommendations based on simplicity and consistency with existing patterns.
