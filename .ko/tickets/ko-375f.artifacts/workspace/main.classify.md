Looking at this ticket, I need to classify the type of work it describes.

## Analysis

The ticket title is: "Migrate agent harnesses from YAML templates to shell scripts"

Key indicators:
- **Action verb**: "Migrate" - this is a concrete directive to change the codebase
- **Specific scope**: Agent harnesses, YAML templates → shell scripts
- **Clear objective**: Transform from one format/implementation to another
- **Expected output**: Code changes (migrated shell scripts to replace YAML templates)

This describes **what to build** - a migration from one technology/format to another. It's not asking "can we do this?" (research), not reporting "something's broken" (bug), and not describing unexpected behavior. It's specifying a concrete refactoring/migration task with clear inputs and outputs.

The work involves:
1. Understanding the current YAML template-based harnesses
2. Converting them to shell scripts
3. Committing the migrated code

This is a **task** - it requires code changes with a clear transformation objective.

```json
{"disposition": "route", "workflow": "task"}
```
