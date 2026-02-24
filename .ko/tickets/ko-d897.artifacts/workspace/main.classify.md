I'll analyze this ticket to determine its classification.

## Ticket Analysis

**Title:** "ko add should accept ticket description via stdin/heredoc or second arg"

**What it asks for:**
- Add support for setting ticket body at creation time
- Three ways to provide description:
  1. `ko add "title" "description text"` — second positional arg
  2. `ko add "title" <<EOF ... EOF` — stdin/heredoc
  3. `ko add "title"` — no description (existing behavior)
- Depends on a rename ticket (ko-90ba)
- Decision clarified: strict priority ordering (stdin > second arg > -d flag)

**Classification reasoning:**

This ticket clearly describes **what to build** — specific features with defined behavior. It's not asking "can we do this?" or reporting unexpected behavior. The requirements are concrete:
- Accept description via multiple input methods
- Implement specific priority ordering
- The work produces committed code changes to the `ko` CLI

The ticket has been through the planning phase already (notes show a plan was created, a decision was clarified). This is ready for implementation.

```json
{"disposition": "route", "workflow": "task"}
```
