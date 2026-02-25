## Goal

Replace YAML-based agent harness templates with executable shell scripts that receive parameters via KO_-namespaced environment variables.

## Context

**Current YAML Architecture:**
- `harness.go` defines `Harness` struct (loaded from YAML) and `TemplateAdapter` (renders YAML templates to exec.Cmd)
- `agent-harnesses/*.yaml` contains built-in harness definitions (embedded via `//go:embed`)
- Harness search order: `.ko/agent-harnesses/` → `~/.config/knockout/agent-harnesses/` → built-ins
- YAML templates support variables like `${prompt}`, `${model}`, `${system_prompt}`, `${allow_all}`, `${allowed_tools}`, `${prompt_with_system}`, `${cursor_allow_all}`
- Template variables expand to multi-argument flags (e.g., `${model}` → `--model\nsonnet`)
- `TemplateAdapter.BuildCommand()` handles template rendering, binary fallback resolution, and stdin setup
- The `AgentAdapter` interface has `BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd`

**Example Current YAML (claude.yaml):**
```yaml
binary: claude
args:
  - "-p"
  - "--output-format"
  - "text"
  - "${allow_all}"
  - "${allowed_tools}"
  - "${model}"
  - "${system_prompt}"
```

**Project Constraints (from INVARIANTS.md):**
- 500-line file limit (harness.go is 173 lines)
- Specs before code (no spec exists yet — needs to be added)
- Zero external runtime dependencies (shell is considered available, like the kernel)

**Testing:**
- `harness_test.go` has comprehensive tests for YAML loading, binary fallback, command building
- `specs/pipeline.feature` documents harness behavior (scenarios 51-75)
- Tests verify template variable expansion, conditional flags, stdin setup, project/user override behavior

**Ticket Context (from notes):**
- **Motivation**: Different agent harnesses have different syntax/affordances; shell scripts ensure invocation intent carries through regardless of harness
- **Architecture choice**: Executable wrapper with KO_-namespaced environment variables
- **Migration strategy**: Replace YAML entirely — convert built-in and user harnesses to shell, deprecate YAML support
- **Binary fallback**: Move to shell scripts using `command -v` or similar

## Approach

Create a new `ShellAdapter` that executes shell scripts instead of rendering YAML templates. Shell harness scripts will be executable files that:
1. Receive parameters via KO_-namespaced environment variables (KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS)
2. Handle binary fallback logic internally using `command -v`
3. Construct and exec the agent CLI command

The Go code will search for shell harnesses (`.ko/agent-harnesses/<name>` as executable file) before falling back to YAML harnesses (for backward compatibility during transition), then remove YAML support entirely. Built-in shell harnesses will be embedded using `//go:embed` as text files, written to a temp location with exec permissions, then executed.

## Tasks

1. **[specs/agent_harnesses.feature]** — Create a new spec file documenting shell harness behavior: environment variable contract, binary fallback expectations, search order, and migration from YAML. Include scenarios for KO_-namespaced env vars, executable detection, binary fallback in shell, and precedence.
   Verify: Spec captures all behavioral requirements.

2. **[agent-harnesses/claude.sh]** — Create shell script version of claude.yaml. Script receives KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS as env vars. Constructs claude CLI args with conditional flags (only add --model if KO_MODEL is set, etc.). Execs claude with prompt via stdin. Include #!/bin/sh shebang.
   Verify: Script is valid shell and follows env var contract.

3. **[agent-harnesses/cursor.sh]** — Create shell script version of cursor.yaml. Handles KO_PROMPT_WITH_SYSTEM instead of separate system prompt (cursor inlines system into prompt). Implements binary fallback using `command -v cursor-agent || command -v agent`. Uses --force instead of --dangerously-skip-permissions for KO_ALLOW_ALL. Passes prompt as -p argument (not stdin).
   Verify: Script handles binary fallback and cursor-specific flags correctly.

4. **[harness.go:ShellAdapter]** — Add new `ShellAdapter` struct that implements `AgentAdapter`. Store the harness script path. Implement `BuildCommand()` to: set KO_* environment variables on exec.Cmd, set KO_PROMPT_WITH_SYSTEM (combining system + prompt), construct KO_ALLOWED_TOOLS as comma-separated string, execute the shell script directly. Return *exec.Cmd pointing to the shell harness.
   Verify: Implements AgentAdapter interface correctly.

5. **[harness.go:LoadHarness]** — Update to search for shell harnesses first. Check for executable file at `<name>` (no extension) before checking `<name>.yaml`. For project and user paths, use os.Stat and check exec bits. For built-in, embed shell scripts alongside YAML and extract to temp location with exec permissions. Return wrapper struct indicating shell vs YAML type.
   Verify: Search order is shell-first, maintains backward compat with YAML.

6. **[harness.go:Harness]** — Add field to distinguish harness type (shell vs YAML). Update parseHarness to handle both. Shell harnesses don't parse YAML, just store script path. Refactor to return `HarnessConfig` struct with `Type` field and union of `ShellPath` or `YAMLConfig`.
   Verify: Both types can be represented and loaded.

7. **[adapter.go:LookupAdapter]** — Update to return `ShellAdapter` when harness type is shell, `TemplateAdapter` when YAML. Check harness type from LoadHarness result and construct appropriate adapter.
   Verify: Correct adapter type is returned for each harness type.

8. **[harness_test.go]** — Add tests for ShellAdapter: test that KO_* env vars are set correctly on exec.Cmd, test that shell script is executed with correct environment, test binary fallback is handled by script (not Go), test that shell harnesses take precedence over YAML. Update existing tests to work with shell harnesses.
   Verify: `go test ./... -run TestShellAdapter` passes, existing harness tests still pass.

9. **[testdata/harness_shell.txtar]** — Add testscript test for shell harness end-to-end: create .ko/agent-harnesses/test-agent script, build a ticket with agent: test-agent, verify script receives correct KO_* env vars. Use a mock script that echoes env vars to verify contract.
   Verify: `go test ./... -run TestScript/harness_shell` passes.

10. **[agent-harnesses/claude.yaml]** — Delete YAML harness (replaced by claude.sh).
    Verify: File deleted, built-in references removed.

11. **[agent-harnesses/cursor.yaml]** — Delete YAML harness (replaced by cursor.sh).
    Verify: File deleted, built-in references removed.

12. **[harness.go:parseHarness]** — Remove YAML parsing logic. Remove yaml.Unmarshal call and YAML struct unmarshaling. Shell harnesses are just paths, no parsing needed beyond checking executability.
    Verify: No more yaml.v3 import or YAML parsing code.

13. **[harness.go:TemplateAdapter]** — Remove TemplateAdapter struct and all methods (BuildCommand, resolveBinary, buildPromptWithSystem). No longer needed since YAML support is removed.
    Verify: TemplateAdapter code fully removed, adapter.go references updated.

14. **[harness.go:embeddedHarnesses]** — Update embed directive from `//go:embed agent-harnesses/*.yaml` to `//go:embed agent-harnesses/*` (will embed .sh files). Update extraction logic to write to temp with exec permissions.
    Verify: Shell scripts are embedded and extractable.

15. **[harness_test.go]** — Remove YAML-specific tests (template expansion, binary_fallbacks field, YAML parsing). Keep tests for shell harnesses, binary fallback in scripts, and search order. Update TestLoadHarness_BuiltInClaude and TestLoadHarness_BuiltInCursor to expect shell harnesses.
    Verify: All tests pass with shell harnesses only.

16. **[README.md or docs]** — Update any documentation mentioning YAML harnesses to reflect shell script approach. Add example shell harness with env var reference. Document migration path for users with custom YAML harnesses.
    Verify: Documentation is accurate and complete.

17. **[build.go:runNode]** — Verify that existing code in runPromptNode already passes allowedTools correctly to BuildCommand. No changes needed, but confirm the call site at line 199 matches the updated contract.
    Verify: Code review confirms correct usage.

## Open Questions

None — all major decisions have been answered:
- Architecture: Executable wrappers with KO_-namespaced env vars (confirmed in ticket notes)
- Migration: Replace YAML entirely, deprecate YAML support (confirmed)
- Binary fallback: Move to shell using `command -v` (confirmed)
- Runtime dependencies: Shell is acceptable (implicit from architecture choice)

One implementation detail to confirm: Should shell harnesses support stdin for prompt (like current claude.yaml with `-p` flag) or always pass via KO_PROMPT env var? Current plan assumes env var only, but scripts can read stdin if needed. The shell script has full control over how it uses the environment.
