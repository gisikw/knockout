## Goal

Replace YAML-based agent harness templates with executable shell scripts for configuring agent CLI invocations.

## Context

The current agent harness system uses YAML configuration files to describe how to invoke different agent CLIs (claude, cursor, etc.):

**Current Architecture:**
- `harness.go` defines the `Harness` struct (loaded from YAML) and `TemplateAdapter` (renders YAML templates to exec.Cmd)
- `agent-harnesses/*.yaml` contains built-in harness definitions (embedded via `//go:embed`)
- Harness search order: `.ko/agent-harnesses/` → `~/.config/knockout/agent-harnesses/` → built-ins
- YAML templates support variables like `${prompt}`, `${model}`, `${system_prompt}`, `${allow_all}`, `${allowed_tools}`, `${prompt_with_system}`, `${cursor_allow_all}`
- Template variables can expand to multi-argument flags (e.g., `${model}` → `--model\nsonnet`)
- The `AgentAdapter` interface has `BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd`
- Two adapter implementations: `TemplateAdapter` (uses YAML harnesses) and `RawCommandAdapter` (legacy backward compat)

**Existing Tests:**
- `harness_test.go` has comprehensive tests for loading harnesses, binary fallback resolution, and command building for both claude and cursor
- Tests verify template variable expansion, conditional flag handling, stdin setup, and project/user override behavior

**Project Constraints (from INVARIANTS.md):**
- 500-line file limit (harness.go is 173 lines, well within limit)
- Specs before code (no spec exists yet for this migration)
- Zero external runtime dependencies (shell scripts would need to be self-contained or require shell interpreter)

## Approach

**The ticket lacks critical context about the migration rationale and design.** Shell scripts could mean:

1. **Executable harness wrappers**: Replace YAML with shell scripts that receive args via env vars and exec the agent CLI
2. **Shell-based templating**: Keep Go orchestration but use shell scripts for template rendering
3. **Full delegation**: Shell scripts handle all command construction and execution

Each has different implications for:
- How parameters (prompt, model, system prompt, allow flags, tools) are passed
- Binary fallback resolution strategy
- Testing approach (shell scripts are harder to unit test than Go code)
- Cross-platform compatibility (bash vs sh vs zsh)
- Runtime dependency posture (violates "zero external runtime dependencies" if shell is required)

Without knowing the design intent, I cannot plan implementation tasks.

## Tasks

Cannot define implementation tasks without understanding:
1. Which shell script approach to use
2. How shell scripts receive parameters (env vars, stdin, args, JSON?)
3. Whether binary fallback logic stays in Go or moves to shell
4. Testing strategy for shell-based harnesses
5. Migration path for existing YAML harnesses (convert automatically? coexist?)

## Open Questions

1. **Design Intent**: What problem does migrating to shell scripts solve? Is this for:
   - Simpler syntax/user experience for custom harnesses?
   - More flexibility in command construction?
   - Better shell pipeline composition?
   - Something else?

2. **Architecture**: Which shell script approach should be used?
   - Option A: Executable scripts that receive params via env vars (PROMPT, MODEL, SYSTEM_PROMPT, ALLOW_ALL, ALLOWED_TOOLS) and exec the agent CLI
   - Option B: Shell scripts as template renderers that output the full command line
   - Option C: Something else?

3. **Interface Contract**: How should shell harness scripts receive input?
   - Environment variables?
   - Arguments?
   - JSON on stdin?
   - A combination?

4. **Binary Resolution**: Should binary fallback logic (currently `binary_fallbacks` in YAML):
   - Stay in Go and pass resolved binary to shell script?
   - Move into shell script (using `command -v` or similar)?

5. **Backward Compatibility**: Should YAML harnesses:
   - Continue to work alongside shell scripts?
   - Be automatically converted to shell scripts?
   - Be deprecated and removed?

6. **Runtime Dependencies**: Shell scripts require a shell interpreter. Does this violate the "zero external runtime dependencies" invariant from INVARIANTS.md? Or is shell considered universally available (like the kernel)?

7. **Testing**: How should shell script harnesses be tested?
   - Keep Go unit tests that shell out to scripts?
   - Add shell script unit tests (bats, shunit2)?
   - Integration tests only?

8. **Migration**: For the built-in harnesses (claude.yaml, cursor.yaml):
   - Should these be converted first as examples?
   - What should the shell script versions look like?

9. **File Naming**: Should shell harnesses be:
   - Named `<agent>.sh` (parallel to current `<agent>.yaml`)?
   - Named something else?
   - Detected by executable bit rather than extension?
