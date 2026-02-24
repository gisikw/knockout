Looking at the ticket title "Migrate agent harnesses from YAML templates to shell scripts", this is clearly asking for a specific code change/migration work.

Key indicators:
- **"Migrate"** - explicit instruction to change implementation from one form to another
- **"from X to Y"** - describes what to build/change, not what's broken
- **Expected output** - migrated code (shell scripts replacing YAML templates)
- **Not investigative** - the problem and solution are both stated; no questions being asked
- **Not a bug** - not reporting unexpected behavior, not a regression

This is a **task**: a concrete implementation job to refactor the agent harnesses codebase.

```json
{"disposition": "route", "workflow": "task"}
```
