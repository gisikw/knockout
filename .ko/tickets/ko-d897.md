---
id: ko-d897
status: open
deps: [ko-90ba]
links: []
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
