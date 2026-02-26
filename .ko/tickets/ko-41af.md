---
id: ko-41af
status: closed
deps: []
created: 2026-02-26T02:52:38Z
type: task
priority: 1
---
# ko serve: ticket-specific commands (close, open, bump, show, update) fail for non-default projects — ticket not found

When ko serve receives a command like close/open/bump/show/update with just a ticket ID (no --project flag), it resolves the ticket path relative to the serve process's cwd. Tickets that live in non-default project dirs (e.g. user tickets in ~/.local/share/punchlist) are not found. The CLI handles this by scanning all registered project dirs, but serve doesn't. Repro: curl -s https://knockout.gisi.network/ko -d '{"argv":["close","user-4ad5"]}' returns ticket not found.
