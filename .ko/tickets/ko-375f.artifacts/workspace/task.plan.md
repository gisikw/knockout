The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-375f.artifacts/plan.md`. 

## Summary

I've updated the plan based on the ticket notes which answered all the open questions from the prior plan:

**Key Decisions (from ticket notes):**
- **Architecture**: Executable wrapper scripts that receive parameters via KO_-namespaced environment variables
- **Migration strategy**: Replace YAML entirely — convert built-in and user harnesses to shell, deprecate YAML support  
- **Binary fallback**: Move logic from Go to shell scripts using `command -v` or similar

**Approach:**
Create a new `ShellAdapter` that executes shell scripts which receive KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, and KO_ALLOWED_TOOLS as environment variables. The shell scripts handle binary fallback internally and construct the appropriate CLI commands.

**Migration Path:**
1. Create shell versions of built-in harnesses (claude.sh, cursor.sh)
2. Add ShellAdapter to execute shell harnesses  
3. Update LoadHarness to search for shell scripts first (backward compat with YAML during transition)
4. Remove all YAML support (TemplateAdapter, parseHarness YAML logic, .yaml files)
5. Update tests and documentation

The plan includes 17 concrete tasks with file paths and verification steps, progressing from specs → implementation → testing → cleanup.
