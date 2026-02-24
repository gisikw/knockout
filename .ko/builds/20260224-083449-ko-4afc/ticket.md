---
id: ko-4afc
status: open
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# ko blocked: surface unblock reasons for blocked tickets

`ko blocked <id>` — output the reason a specific ticket is blocked.
`ko blocked` — output reasons for all blocked tickets.

Consider: blocking a ticket should always require a reason. Maybe enforce
that the reason lands in frontmatter or a structured field so it's easy
to extract programmatically, not just grep through notes.
