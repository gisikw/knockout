## Goal
Replace YAML-based agent harness templates with executable shell scripts.

## Context
Agent harnesses currently use YAML files with string-substitution templates (`${variable}` syntax) to build CLI commands for agents like claude and cursor. These live in:
- `agent-harnesses/claude.yaml`
- `agent-harnesses/cursor.yaml`

The harness loading mechanism (harness.go:26-50) searches in this order:
1. `.ko/agent-harnesses/<name>.yaml`
2. `~/.config/knockout/agent-harnesses/<name>.yaml`
3. Embedded built-ins (from `agent-harnesses/*.yaml`)

The `TemplateAdapter` (harness.go:60-147) implements the current template rendering:
- Builds a vars map with prompt, model, system_prompt, allow_all, etc.
- Does string replacement on the Args array
- Handles conditional flags (empty vars are omitted)
- Splits flag variables on newlines for multi-arg expansion

This approach is limited compared to shell scripts, which can:
- Use native shell conditionals and logic
- Access shell utilities for complex transformations
- Be more familiar to users customizing harnesses
- Eliminate custom template parsing code

The migration should maintain backward compatibility where feasible and preserve all existing test coverage.

## Approach
Replace the YAML template system with shell scripts that output command arguments as newline-separated strings. The scripts will receive harness parameters as environment variables and output the full command line (binary + args). The Go code will execute the script and parse its output to build the exec.Cmd.

## Tasks
1. [agent-harnesses/claude.yaml → agent-harnesses/claude.sh] — Convert the claude harness YAML template to a shell script. The script receives KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS as env vars. Output format: first line is the binary path, subsequent lines are arguments (one per line). Empty variables should result in omitted flags.
   Verify: `./agent-harnesses/claude.sh` is executable and produces expected output when env vars are set.

2. [agent-harnesses/cursor.yaml → agent-harnesses/cursor.sh] — Convert the cursor harness YAML template to a shell script. Handle binary_fallbacks by trying cursor-agent first, then agent, outputting the first found. Inline system prompt into the main prompt (KO_PROMPT_WITH_SYSTEM variable).
   Verify: `./agent-harnesses/cursor.sh` is executable and handles binary fallbacks correctly.

3. [harness.go:Harness struct] — Replace YAML-based Harness struct with a simpler structure that just points to a shell script path. Remove Binary, BinaryFallbacks, and Args fields. Add ScriptPath string field.
   Verify: Code compiles with new struct definition.

4. [harness.go:LoadHarness] — Update LoadHarness to search for `.sh` files instead of `.yaml` files. Update filename from `name + ".yaml"` to `name + ".sh"`. Update embed directive to `//go:embed agent-harnesses/*.sh`.
   Verify: LoadHarness("claude") finds claude.sh instead of claude.yaml.

5. [harness.go:TemplateAdapter → ScriptAdapter] — Replace TemplateAdapter with ScriptAdapter that executes the shell script and parses its output. The adapter's BuildCommand method should: set env vars (KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS, KO_PROMPT_WITH_SYSTEM), execute the script, parse output (first line = binary, remaining lines = args), build exec.Cmd. Remove all template rendering logic (resolveBinary, vars map, string substitution).
   Verify: Code compiles and adapter interface is preserved.

6. [harness_test.go] — Update all tests to work with shell scripts. TestLoadHarness_BuiltInClaude should verify claude.sh loads. TestLoadHarness_UserOverride should write a test .sh file instead of .yaml. TestTemplateAdapter_ClaudeCommand should become TestScriptAdapter_ClaudeCommand and verify the script execution produces correct args. Add tests for script parsing edge cases (empty output, malformed output, script errors).
   Verify: `go test ./... -run TestLoadHarness` passes; `go test ./... -run TestScriptAdapter` passes.

7. [README.md:290-335] — Update the "Custom Agent Harnesses" documentation section. Replace YAML template examples with shell script examples. Update file extension references from `.yaml` to `.sh`. Document the env vars available to scripts (KO_PROMPT, KO_MODEL, etc.) and the expected output format (first line = binary, subsequent lines = args). Remove references to YAML template syntax.
   Verify: Documentation accurately describes the new shell script approach.

8. [harness.go:parseHarness] — Remove parseHarness function and yaml.v3 import. No longer needed since scripts don't require YAML parsing.
   Verify: Code compiles without YAML dependency; `go mod tidy` removes gopkg.in/yaml.v3 if unused elsewhere.

9. [specs/pipeline.feature:51-76] — Update specs referencing custom harnesses to use `.sh` extension instead of `.yaml`. Update scenario descriptions if needed to reflect shell script approach.
   Verify: Specs accurately describe expected behavior with shell scripts.

10. [Build and test] — Run full test suite to verify all changes work together. Test custom harness loading from all three locations (project, user, built-in). Test both claude and cursor harnesses produce correct commands.
    Verify: `go test ./...` passes; `ko agent build` works with both built-in harnesses.

## Open Questions
1. **Backward compatibility**: Should we maintain support for YAML harnesses during a deprecation period, or make a clean break? The ticket says "migrate" which suggests replacing entirely, but user-custom YAML harnesses in `~/.config/` would break immediately. Decision needed on whether to: (a) support both formats with .sh taking precedence, (b) migrate immediately and communicate breaking change, or (c) provide a migration tool.

2. **Error handling**: How should the system handle shell script failures (non-zero exit, no output, malformed output)? Should these fall through to the next harness location in the search order, or fail immediately with a clear error?

3. **Security**: Shell scripts executing arbitrary code is more powerful but potentially riskier than declarative YAML. Should there be any sandboxing, validation, or warnings when using custom harnesses? Or is this acceptable given harnesses are already trusted code?

4. **Script permissions**: The plan assumes scripts will be executable (chmod +x). Should LoadHarness check and set execute permissions automatically, or require them to be pre-set? Embedded scripts might need special handling.
