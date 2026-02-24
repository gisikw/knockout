Looking at this ticket, I need to classify it based on what type of work it requires.

**Ticket Analysis:**

The ticket describes a specific refactoring task:
- **What to do**: Consolidate `.ko/` top-level files into a single `.ko/config.yaml`
- **Scope**: Merge `pipeline.yml` and other project-level config files
- **Purpose**: Create a unified configuration location for pipeline config, agent harness overrides, and project-level settings
- **Context**: This is a prerequisite for a larger feature (agent harness extraction)

**Classification Decision:**

This is a **task** ticket. It:
1. Asks for specific code/configuration changes (consolidation of files)
2. Has a clear architectural goal (single config file)
3. Describes *what to build* (unified config structure) rather than reporting broken behavior
4. The expected output is committed changes to the configuration structure

```json
{"disposition": "route", "workflow": "task"}
```
