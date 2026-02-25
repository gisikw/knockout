Looking at this ticket, I need to classify what kind of work it requires.

The ticket clearly describes **what to build**: it's asking for a systematic update to the codebase to use new CLI syntax that has already been established. It provides:
- Specific locations to audit and update (.ko/prompts/, cmd_build.go, cmd_serve.go, etc.)
- A clear mapping of old command names to new ones
- The expected outcome: updating all references to use the new syntax

This is not asking "should we do this?" or "why is something broken?" — it's asserting that the agent build pipeline needs to be updated to reflect the CLI simplification that has already been completed, with specific guidance on what needs to change.

This is a **task**: the expected output is committed code changes that update the CLI references throughout the codebase.

```json
{"disposition": "route", "workflow": "task"}
```
