I'll analyze this ticket to classify it appropriately.

**Ticket Analysis:**

The ticket clearly describes **what to build**:
- "Support inline prompt text in pipeline nodes"
- Current state: requires `prompt: filename.md` loading from `.ko/prompts/`
- Desired state: support inline prompt text directly in config using YAML multiline syntax
- Clear implementation guidance: "When `prompt:` value contains whitespace/newlines, treat it as inline text. When it looks like a filename (no spaces, ends in .md), load from file as today."

This is a **feature request** with:
- Specific requirements defined
- Clear acceptance criteria (handle both inline and file-based prompts)
- A concrete example of the desired syntax
- Dependencies noted (prereq for ko-4895)

This is straightforward implementation work, not investigation or diagnosis.

```json
{"disposition": "route", "workflow": "task"}
```
