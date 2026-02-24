---
id: ko-d897
status: open
deps: [ko-90ba]
created: 2026-02-24T07:30:19Z
type: task
priority: 2
---
# ko add should accept ticket description via stdin/heredoc or second arg

Currently `ko create` (soon `ko add`) only takes a title. Support setting
the ticket body at creation time:

- `ko add "title" "description text"` — second positional arg
- `ko add "title" <<EOF ... EOF` — stdin/heredoc (same pattern as ko note)
- `ko add "title"` — no description, same as today

Depends on ko-90ba (rename create to add) — implement against the new name.

## Notes

**2026-02-24 07:40:00 UTC:** ko: FAIL at node 'actionable' — Plan contains an open question that requires clarification: Should description sources have strict priority (stdin wins over arg, wins over -d) OR should they all merge together? The plan assumes merge behavior but notes this decision should be confirmed. Ticket author must clarify the intended precedence/merge strategy before implementation can proceed.

**2026-02-24 13:02:31 UTC:** Decision: strict priority, no merging. stdin > second positional arg > -d flag.

**2026-02-24 13:06:01 UTC:** ko: FAIL at node 'classify' — node 'classify' failed after 3 attempts: disposition missing required 'disposition' field
