## Analysis

The plan artifact clearly documents the current architecture and identifies critical missing decisions. Let me assess against the three criteria:

### 1. **Open Questions Section**
The plan explicitly lists 9 open questions spanning:
- **Design Intent** (Q1): What problem does migrating to shell scripts solve?
- **Architecture** (Q2): Which approach should be used?
- **Interface Contract** (Q3): How should shell scripts receive parameters?
- **Binary Resolution** (Q4): Where should binary fallback logic live?
- **Backward Compatibility** (Q5): Should YAML harnesses coexist or be replaced?
- **Runtime Dependencies** (Q6): Does shell violate the "zero external runtime dependencies" invariant?
- **Testing Strategy** (Q7): How to test shell scripts?
- **Migration of Built-ins** (Q8): How to convert existing harnesses?
- **File Naming** (Q9): Naming convention for shell harnesses?

These are genuine architectural decisions that cannot be inferred from the codebase alone—they require human judgment about design intent and trade-offs.

### 2. **Completeness**
The plan has:
- ✓ Clear goal statement
- ✓ Thorough context (current architecture, tests, constraints)
- ✗ No ordered implementation tasks (blocked by open questions)
- ✗ No verification steps (deferred until questions are answered)

The plan correctly acknowledges incompleteness—it ends by stating "Cannot define implementation tasks without understanding..." the key design decisions.

### 3. **Feasibility**
The plan references plausible files:
- `harness.go` (173 lines, exists)
- `harness_test.go` (exists)
- `agent-harnesses/*.yaml` (referenced as embedded)
- Constraints from `INVARIANTS.md` (mentioned)

No red flags on file existence or patterns.

## Conclusion

This plan **needs human input** to proceed. The decision node should collect answers to the key architectural questions before implementation can be planned. I'll structure the 9 questions into a consolidated set focusing on the most critical decisions:

```json
{
  "disposition": "needs_input",
  "plan_questions": [
    {
      "id": "q1",
      "question": "What is the primary motivation for migrating from YAML to shell scripts?",
      "context": "Understanding the design intent will inform all downstream decisions.",
      "options": [
        {
          "label": "Simpler syntax for custom harnesses",
          "value": "simpler_syntax",
          "description": "Shell scripts are more intuitive for users to write than YAML templates"
        },
        {
          "label": "More flexible command construction",
          "value": "flexible_construction",
          "description": "Shell allows dynamic logic, conditionals, and piping that YAML templates cannot express"
        },
        {
          "label": "Better shell pipeline composition",
          "value": "pipeline_composition",
          "description": "Shell scripts can be composed into pipelines more naturally"
        },
        {
          "label": "Other (please specify)",
          "value": "other",
          "description": "Different rationale"
        }
      ]
    },
    {
      "id": "q2",
      "question": "Which architecture should the implementation use?",
      "context": "Option A: Shell scripts are invoked as executables with parameters passed via environment variables. Option B: Shell scripts are renderers that output a complete command line for Go to execute. Option C: Shell scripts handle all orchestration including binary resolution and execution.",
      "options": [
        {
          "label": "Option A: Executable wrapper scripts with env vars",
          "value": "env_var_wrapper",
          "description": "Go passes params (PROMPT, MODEL, etc.) via env vars to shell scripts that exec the agent CLI"
        },
        {
          "label": "Option B: Shell template renderers",
          "value": "template_renderer",
          "description": "Shell scripts output the full command line; Go then executes it"
        },
        {
          "label": "Option C: Full delegation to shell",
          "value": "full_delegation",
          "description": "Shell scripts handle binary resolution, parameter processing, and execution"
        }
      ]
    },
    {
      "id": "q3",
      "question": "Should YAML harnesses continue to work after migration?",
      "context": "The plan mentions INVARIANTS.md constraint about 'zero external runtime dependencies'. Shell scripts introduce a runtime dependency on a shell interpreter, unless we consider shell universally available.",
      "options": [
        {
          "label": "Coexist: Support both YAML and shell scripts",
          "value": "coexist",
          "description": "Both formats work; search order defines priority (e.g., shell scripts first, then YAML)"
        },
        {
          "label": "Replace: Convert YAML to shell and deprecate YAML",
          "value": "replace",
          "description": "Migrate built-in and user harnesses to shell; YAML support is removed"
        },
        {
          "label": "Auto-convert: YAML files are automatically converted to shell",
          "value": "auto_convert",
          "description": "Tooling converts YAML harnesses to shell scripts; old files may remain but are no longer used"
        }
      ]
    },
    {
      "id": "q4",
      "question": "Should binary fallback logic stay in Go or move to shell?",
      "context": "Currently YAML harnesses support binary_fallbacks (e.g., try 'claude' then 'claude-4.5'). This can be implemented in Go (before calling shell) or in the shell script itself.",
      "options": [
        {
          "label": "Stay in Go: Resolve binary, pass to shell",
          "value": "go_resolution",
          "description": "Go determines the final binary path and passes it to the shell script"
        },
        {
          "label": "Move to shell: Shell handles fallback logic",
          "value": "shell_resolution",
          "description": "Shell scripts use command -v or similar to find the binary"
        }
      ]
    }
  ]
}
```
