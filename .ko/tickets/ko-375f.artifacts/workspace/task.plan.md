Perfect! I've created a comprehensive plan for migrating agent harnesses from YAML templates to shell scripts. The plan includes:

## Summary

**Goal**: Replace YAML-based agent harnesses with executable shell scripts for simpler command construction.

**Key Findings**:
- Agent harnesses currently use YAML templates (claude.yaml, cursor.yaml) with template variables
- These are parsed and rendered through TemplateAdapter in harness.go
- The system already has an AgentAdapter interface that abstracts the implementation

**Approach**: 
Convert YAML templates to shell scripts that receive env vars and output command lines. This eliminates complex YAML parsing logic and makes harnesses directly inspectable as shell code.

**10 Tasks** outlined covering:
1-2. Converting claude.yaml and cursor.yaml to .sh scripts
3-6. Updating Go code (LoadHarness, Harness struct, adapters)
7. Updating tests
8. Updating documentation
9-10. Updating go:embed and cleaning up old YAML files

**6 Open Questions** identified that need human input:
1. Script output format (args vs full command line, quoting/escaping)
2. Binary resolution strategy (script vs Go responsibility)
3. Backward compatibility with YAML during transition
4. Cross-platform support (Windows .bat/.ps1 vs Unix .sh)
5. Stdin handling mechanism in script output
6. Error handling strategy (fallback vs fail-hard)

These questions represent genuine architectural decisions that need product input before implementation can proceed.
