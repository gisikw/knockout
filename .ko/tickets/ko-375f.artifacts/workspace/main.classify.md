Looking at this ticket, I need to classify the type of work it requires.

**Analysis:**

The ticket is titled "Migrate agent harnesses from YAML templates to shell scripts" and the notes section clearly shows:
- The motivation is established (ensure agent invocation intent carries through regardless of harness syntax)
- The architecture is decided (executable wrapper with KO_ namespaced env vars)
- The approach is determined (replace YAML with shell scripts, deprecate YAML)
- Implementation details are clear (shell handles fallback logic using command -v)

This is a **task ticket** because:
1. It explicitly describes what to build: migrate YAML templates to shell scripts
2. The expected output is committed code implementing this migration
3. The architectural decisions have been made and answered in the notes
4. It's not asking "can we do this?" or "what would it take?" - those questions have been answered
5. It's not reporting unexpected behavior - it's requesting a specific implementation

The ticket has moved past the research/design phase and is ready for implementation work.

```json
{"disposition": "route", "workflow": "task"}
```
