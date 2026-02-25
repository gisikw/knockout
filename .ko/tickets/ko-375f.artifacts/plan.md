## Goal
Replace YAML-based agent harnesses with executable shell scripts for simpler, more direct command construction.

## Context
Agent harnesses currently use YAML templates (claude.yaml, cursor.yaml in agent-harnesses/) that are parsed by harness.go and rendered through TemplateAdapter. The YAML format includes:
- `binary` or `binary_fallbacks` fields for executable resolution
- `args` array with template variables like `${prompt}`, `${model}`, `${system_prompt}`, `${allow_all}`, `${allowed_tools}`
- Special handling for stdin detection (standalone `-p` flag)
- Conditional flag expansion (newline-based for non-prompt variables)

Key files:
- harness.go:14-15 — embeds YAML files via go:embed
- harness.go:24-50 — LoadHarness searches project/user/embedded locations
- harness.go:60-173 — TemplateAdapter renders YAML templates to exec.Cmd
- adapter.go:16-23 — LookupAdapter delegates to harness loading
- build.go:447-448 — Pipeline uses p.Adapter().BuildCommand() for prompt nodes
- pipeline.go:30-43 — Pipeline.Adapter() returns AgentAdapter interface

Tests in harness_test.go verify built-in harnesses, user/project overrides, binary fallback resolution, and template rendering.

## Approach
Replace YAML templates with shell scripts that output the command to execute. Harnesses become executable scripts that receive environment variables (MODEL, SYSTEM_PROMPT, ALLOW_ALL, etc.) and write the command line to stdout. The Go code will parse the script output and build exec.Cmd from it.

This simplifies the system by:
1. Eliminating complex YAML parsing and template rendering logic
2. Making harnesses directly inspectable as shell code
3. Allowing harness authors to use standard shell tools for conditional logic
4. Reducing the Go code surface area (no more TemplateAdapter)

## Tasks
1. [agent-harnesses/claude.yaml → agent-harnesses/claude.sh] — Convert Claude harness to shell script format. Script reads env vars (PROMPT, MODEL, SYSTEM_PROMPT, ALLOW_ALL, ALLOWED_TOOLS) and outputs command line. Handle stdin via special marker in output.
   Verify: Script produces correct command line for various input combinations.

2. [agent-harnesses/cursor.yaml → agent-harnesses/cursor.sh] — Convert Cursor harness to shell script with binary fallback logic (try cursor-agent, then agent). Use prompt_with_system pattern (inline system prompt into user prompt).
   Verify: Script produces correct command line and handles binary fallbacks.

3. [harness.go:LoadHarness] — Update to search for .sh files instead of .yaml. Remove YAML parsing (parseHarness function). Add script execution logic that passes env vars and captures stdout.
   Verify: LoadHarness returns executable path and env var template.

4. [harness.go:Harness struct] — Simplify struct to hold just the script path. Remove Binary, BinaryFallbacks, Args fields that were YAML-specific.
   Verify: Struct matches new script-based approach.

5. [harness.go:TemplateAdapter → ScriptAdapter] — Replace TemplateAdapter with ScriptAdapter that executes the harness script with appropriate env vars, parses the output, and builds exec.Cmd. Handle stdin detection from script output.
   Verify: Adapter builds correct commands for both Claude and Cursor cases.

6. [adapter.go:LookupAdapter] — Update to create ScriptAdapter instead of TemplateAdapter. No other changes needed since it returns AgentAdapter interface.
   Verify: LookupAdapter continues to work with new adapter type.

7. [harness_test.go] — Update all tests to work with shell script harnesses instead of YAML. Test script execution, command building, user/project overrides, and error cases.
   Verify: `go test -run TestLoad ./...` and `go test -run TestTemplate ./...` pass.

8. [README.md:290-335] — Update Custom Agent Harnesses section to document shell script format instead of YAML. Show example harness scripts with env var handling and output format.
   Verify: Documentation accurately reflects new shell script approach.

9. [harness.go:14-15] — Update go:embed directive to embed *.sh files instead of *.yaml. Clean up old YAML imports (gopkg.in/yaml.v3).
   Verify: Build succeeds with embedded shell scripts.

10. [agent-harnesses/*.yaml] — Delete old YAML harness files after verifying shell script versions work.
    Verify: Built-in harnesses resolve correctly via embedded *.sh files.

## Open Questions
1. **Output format for scripts**: Should scripts output just the command args (one per line), or the full command line? If full command line, how do we handle:
   - Quoting and escaping (e.g., prompts with newlines or special chars)?
   - Binary path vs args separation?
   - Stdin marker (how does the script signal "read prompt from stdin")?

2. **Binary resolution**: Should the script output the binary path, or should Go still handle binary resolution? Current YAML approach has binary_fallbacks — should scripts return multiple candidates, or just the first available?

3. **Backward compatibility**: Should we support YAML harnesses alongside shell scripts during a transition period, or cleanly break with YAML? Affects search order in LoadHarness.

4. **Platform support**: Shell scripts assume Unix/bash. Windows users would need .bat/.ps1 equivalents. Should we:
   - Require WSL/bash on Windows?
   - Support multiple script types (.sh, .bat, .ps1)?
   - Keep YAML for Windows compatibility?

5. **Stdin handling**: Current YAML approach detects `-p` flag without a template variable following it. How should scripts signal stdin mode? Special output prefix like `STDIN:true`? Separate line in output?

6. **Error handling**: If a harness script fails or produces invalid output, should we fall back to a default harness, or fail hard with a clear error message?
