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
- 500-line file limit (harness.go is currently 173 lines)
- Specs before code (no spec exists yet — needs to be added)
- Zero external runtime dependencies (shell is considered available, like the kernel)
- Every behavior has a spec in `specs/*.feature` and corresponding test in `testdata/*.txtar`

**Testing:**
- `harness_test.go` has comprehensive tests for YAML loading, binary fallback, command building
- `specs/pipeline.feature` documents harness behavior (scenarios 51-75)
- Tests verify template variable expansion, conditional flags, stdin setup, project/user override behavior
- Testscript tests use `testdata/*.txtar` format

**Ticket Context (from notes):**
- **Motivation**: Different agent harnesses have different syntax/affordances; shell scripts ensure invocation intent carries through regardless of harness
- **Architecture choice**: Executable wrapper with KO_-namespaced environment variables
- **Migration strategy**: Replace YAML entirely — convert built-in and user harnesses to shell, deprecate YAML support
- **Binary fallback**: Move to shell scripts using `command -v` or similar
- **Prompt passing**: Environment variable only (KO_PROMPT). All scripts receive prompt via KO_PROMPT environment variable and construct their own arguments.

## Approach

Create a new `ShellAdapter` that executes shell scripts instead of rendering YAML templates. Shell harness scripts will be executable files that:
1. Receive parameters via KO_-namespaced environment variables (KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS)
2. Handle binary fallback logic internally using `command -v`
3. Construct and exec the agent CLI command with the prompt passed however that agent expects it (stdin, argument, etc.)

The Go code will search for shell harnesses (`.ko/agent-harnesses/<name>` as executable file) before falling back to YAML harnesses (for backward compatibility during transition), then remove YAML support entirely. Built-in shell harnesses will be embedded using `//go:embed` as text files, written to a temp location with exec permissions, then executed.

## Tasks

1. **[specs/agent_harnesses.feature]** — Create a new spec file documenting shell harness behavior: environment variable contract (KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS), binary fallback expectations, search order, and migration from YAML. Include scenarios for KO_-namespaced env vars, executable detection, binary fallback in shell, and precedence (project > user > built-in).
   Verify: Spec captures all behavioral requirements.

2. **[agent-harnesses/claude.sh]** — Create shell script version of claude.yaml. Script receives KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS as env vars. Constructs claude CLI args with conditional flags (only add --model if KO_MODEL is set, etc.). Passes prompt via stdin (echo "$KO_PROMPT" | claude ...). Include #!/bin/sh shebang.
   Verify: Script is valid shell and follows env var contract.

3. **[agent-harnesses/cursor.sh]** — Create shell script version of cursor.yaml. Combines KO_SYSTEM_PROMPT and KO_PROMPT into single prompt for -p argument (cursor doesn't support separate system prompt). Implements binary fallback using `command -v cursor-agent || command -v agent`. Uses --force instead of --dangerously-skip-permissions for KO_ALLOW_ALL. Passes prompt as -p argument (not stdin).
   Verify: Script handles binary fallback and cursor-specific flags correctly.

4. **[harness.go:ShellAdapter]** — Add new `ShellAdapter` struct that implements `AgentAdapter`. Store the harness script path. Implement `BuildCommand()` to: set KO_* environment variables on exec.Cmd (KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL as "true"/"false", KO_ALLOWED_TOOLS as comma-separated string), execute the shell script directly. Return *exec.Cmd pointing to the shell harness.
   Verify: Implements AgentAdapter interface correctly.

5. **[harness.go:LoadHarness]** — Update to search for shell harnesses first. Check for executable file at `<name>` (no extension) before checking `<name>.yaml`. For project and user paths, use os.Stat and check exec bits. For built-in, embed shell scripts alongside YAML and extract to temp location with exec permissions (os.WriteFile with 0755). Return wrapper struct indicating shell vs YAML type.
   Verify: Search order is shell-first, maintains backward compat with YAML.

6. **[harness.go:Harness]** — Refactor into `HarnessConfig` struct with `Type` field (shell or yaml) and union of `ScriptPath` (for shell) or `YAMLTemplate` (for YAML). Update parseHarness to handle both. Shell harnesses just store script path, YAML harnesses parse into existing Harness struct.
   Verify: Both types can be represented and loaded.

7. **[adapter.go:LookupAdapter]** — Update to return `ShellAdapter` when harness type is shell, `TemplateAdapter` when YAML. Check harness type from LoadHarness result and construct appropriate adapter.
   Verify: Correct adapter type is returned for each harness type.

8. **[harness_test.go]** — Add tests for ShellAdapter: test that KO_* env vars are set correctly on exec.Cmd, test that shell script path is used as command, test that shell harnesses take precedence over YAML when both exist. Update existing tests to work with shell harnesses (TestLoadHarness_BuiltInClaude, TestLoadHarness_BuiltInCursor should expect shell not YAML).
   Verify: `go test ./... -run TestShellAdapter` passes, existing harness tests still pass.

9. **[testdata/agent_harnesses/shell_harness.txtar]** — Add testscript test for shell harness end-to-end: create .ko/agent-harnesses/test-agent script that echoes env vars to a file, create pipeline with agent: test-agent, build a ticket with a prompt node, verify script received correct KO_* env vars (check the file contents). Use a mock script for verification.
   Verify: `go test ./... -run TestScript/agent_harnesses/shell_harness` passes.

10. **[agent-harnesses/claude.yaml]** — Delete YAML harness (replaced by claude.sh). Update embed directive to exclude .yaml files.
    Verify: File deleted, built-in references removed.

11. **[agent-harnesses/cursor.yaml]** — Delete YAML harness (replaced by cursor.sh).
    Verify: File deleted.

12. **[harness.go:parseHarness]** — Remove YAML parsing logic. Remove yaml.Unmarshal call and YAML struct unmarshaling. Shell harnesses are just paths, no parsing needed beyond checking executability.
    Verify: No more yaml.v3 import or YAML parsing code in parseHarness.

13. **[harness.go:TemplateAdapter]** — Remove TemplateAdapter struct and all methods (BuildCommand, resolveBinary, buildPromptWithSystem). No longer needed since YAML support is removed.
    Verify: TemplateAdapter code fully removed, adapter.go references updated.

14. **[harness.go:embeddedHarnesses]** — Update embed directive from `//go:embed agent-harnesses/*.yaml` to `//go:embed agent-harnesses/*.sh` (only embed shell scripts). Update extraction logic to write to temp with exec permissions (0755).
    Verify: Shell scripts are embedded and extractable, YAML files no longer embedded.

15. **[harness_test.go]** — Remove YAML-specific tests (template expansion, binary_fallbacks field, YAML parsing, TestTemplateAdapter_* tests). Keep tests for shell harnesses, search order, and precedence. Update remaining tests to expect shell harnesses only.
    Verify: All tests pass with shell harnesses only, no YAML tests remain.

16. **[README.md:Custom Agent Harnesses section]** — Update documentation to reflect shell script approach. Replace YAML examples with shell script examples showing KO_* env vars. Document that scripts receive KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS and must construct their own command invocation. Add migration guide for users with custom YAML harnesses (convert to shell).
    Verify: Documentation is accurate and complete.

17. **[build.go:runNode]** — Verify that existing code at line 199 (runNode call) passes allowedTools correctly to BuildCommand. No changes needed to call site, but confirm the signature matches and parameters are passed through correctly.
    Verify: Code review confirms correct usage.

## Open Questions

None — all major decisions have been answered in the ticket notes:
- Architecture: Executable wrappers with KO_-namespaced env vars (confirmed)
- Migration: Replace YAML entirely, deprecate YAML support (confirmed)
- Binary fallback: Move to shell using `command -v` (confirmed)
- Prompt passing: Environment variable only (KO_PROMPT) — confirmed in latest ticket note
- Runtime dependencies: Shell is acceptable (implicit from architecture choice)

The implementation detail about prompt passing has been resolved: scripts receive prompts via KO_PROMPT environment variable and decide how to pass it to the agent (stdin, argument, etc.).
