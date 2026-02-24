## Goal
Replace YAML-based agent harness templates with executable shell scripts.

## Context
The harness system was recently extracted from hardcoded Go (ko-1930) into YAML templates. The current implementation:

- **harness.go** (lines 14-173): Defines the `Harness` struct with YAML tags, embeds `agent-harnesses/*.yaml` files, and implements `TemplateAdapter` that renders YAML templates by replacing `${variable}` placeholders and building exec.Cmd
- **agent-harnesses/claude.yaml** and **cursor.yaml**: YAML configs describing binary name, args array with template variables like `${prompt}`, `${model}`, `${allow_all}`
- **adapter.go**: Defines the `AgentAdapter` interface and `LookupAdapter` function that loads harnesses
- **build.go** (line 418): Calls `p.Adapter().BuildCommand()` to create the agent command

The YAML approach has limitations:
- Complex template rendering logic in Go (conditional flag expansion, stdin detection)
- Hard to express conditional logic or complex argument building
- Requires Go code changes to add new template features

Shell scripts would be more flexible:
- Native conditional logic, string manipulation
- No template rendering needed in Go
- Users can extend harnesses without understanding Go code
- Direct execution model (script receives env vars, produces command)

## Approach
Replace the YAML template system with executable shell scripts that output command-line invocations. The Go code will execute the harness script, pass context via environment variables, and capture the script's output as the command to run.

Harness scripts receive inputs via env vars (`KO_PROMPT`, `KO_MODEL`, `KO_SYSTEM_PROMPT`, `KO_ALLOW_ALL`, `KO_ALLOWED_TOOLS`), and output the final command line to stdout. The Go harness loader executes the script and parses the output to build an exec.Cmd.

## Tasks
1. [harness.go] — Replace the YAML-based Harness struct and TemplateAdapter with a ShellHarness struct that executes a shell script. Remove the `embed.FS` for YAML files, parseHarness function, and template rendering logic. Add a new adapter that runs the shell script with env vars and parses stdout to build exec.Cmd.
   Verify: Existing tests pass (will need updates to match new interface).

2. [harness.go:LoadHarness] — Update the function to look for `.sh` files instead of `.yaml` files in the search path (project `.ko/agent-harnesses/`, user `~/.config/knockout/agent-harnesses/`, built-in embedded scripts).
   Verify: LoadHarness("claude") and LoadHarness("cursor") return shell-based harnesses.

3. [agent-harnesses/claude.yaml → claude.sh] — Convert the YAML template to a shell script that reads env vars and outputs a command line. Handle stdin mode (prompt via stdin with `-p` flag) and conditional flags (model, system prompt, allow_all, allowed_tools).
   Verify: Script outputs correct command for various input combinations.

4. [agent-harnesses/cursor.yaml → cursor.sh] — Convert the YAML template to a shell script. Handle binary fallback resolution (try cursor-agent, then agent), inline system prompt into user prompt, use cursor-specific flags (--force instead of --dangerously-skip-permissions).
   Verify: Script outputs correct command with fallback binary resolution and inlined system prompt.

5. [harness.go] — Update the embed directive from `agent-harnesses/*.yaml` to `agent-harnesses/*.sh` to embed the new shell scripts.
   Verify: Built binary contains embedded shell scripts, not YAML files.

6. [harness_test.go] — Update all tests to work with shell-based harnesses. Tests should verify that scripts output correct commands for various input combinations (with/without model, with/without system prompt, allow_all true/false, etc.).
   Verify: `go test ./... -run TestLoadHarness -v` passes.

7. [harness_test.go] — Remove or update tests specific to YAML parsing (parseHarness, template variable rendering). Add tests for shell script execution and command parsing.
   Verify: `go test ./...` passes with no YAML-specific test failures.

8. [README.md:285-328] — Update the "Custom Agent Harnesses" section documentation to reflect shell script approach instead of YAML. Update examples to show shell script format with env vars instead of YAML with template variables.
   Verify: Documentation accurately describes the new shell-based harness system.

## Open Questions
1. **Script output format**: How should the shell script communicate the command to Go? Options:
   - Output the full command line to stdout (e.g., `claude -p --output-format text --model sonnet`), and Go parses it into args
   - Output only the args (one per line), and Go knows the binary from a separate mechanism
   - Use a structured format (JSON?) for more complex needs (stdin handling, env vars)

   **Recommendation**: Full command line to stdout, with special handling for stdin. If the output starts with `@stdin:`, the rest is the command, and Go should pass prompt via stdin. Otherwise, the prompt is already incorporated as an arg.

2. **Binary resolution**: In YAML, binary_fallbacks allowed trying multiple binaries. How do shell scripts handle this?
   - Script contains its own fallback logic (using `command -v` or similar)
   - Script outputs just args, Go resolves binary from a metadata comment in the script

   **Recommendation**: Scripts handle their own binary resolution using `command -v`. This gives maximum flexibility.

3. **Error handling**: What if the harness script fails or outputs malformed data?
   - Return error with script stderr
   - Fall back to a default adapter

   **Recommendation**: Return error with script stderr. No silent fallbacks—if a harness script is broken, the user needs to know.

4. **Backward compatibility**: Should we support both YAML and shell harnesses during a transition period?
   - Support both, check for .sh first then .yaml
   - Hard cutover, remove YAML support entirely

   **Recommendation**: Hard cutover. The YAML system was just added in ko-1930 and hasn't shipped widely. A clean break is simpler than maintaining two systems.
