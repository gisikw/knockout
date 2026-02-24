I'll analyze this ticket to classify it.

**Ticket Analysis:**

The ticket "ko add should accept ticket description via stdin/heredoc or second arg" is requesting a **feature enhancement**. 

Key indicators:
- **Clear specification**: The ticket explicitly describes what to build with three specific implementation patterns:
  - `ko add "title" "description text"` — second positional arg
  - `ko add "title" <<EOF ... EOF` — stdin/heredoc 
  - `ko add "title"` — backwards compatible, no description
- **Describes the desired output**: The expected result is committed code changes to the `ko add` command
- **Depends on a known task**: It references ko-90ba (rename create to add) which is already tracked
- **Not asserting breakage**: It's not reporting that something is broken; it's requesting new functionality

**Classification: TASK**

This is clearly a task ticket because it specifies what to build (feature implementation) rather than reporting what's broken or asking investigative questions.

```json
{"disposition": "route", "workflow": "task"}
```
