---
id: ko-5734
status: open
deps: []
created: 2026-02-25T16:01:32Z
type: task
priority: 3
---
# Remove ko query command (redundant with --json on list commands)

ko query outputs all tickets as JSONL with no filtering. Now that ko ls, ready, blocked, closed, and resolved all support --json, query is redundant. The list commands collectively cover every status. Remove the command, its tests, and the serve whitelist entry. Note: ko serve SSE uses ListTickets() directly, not the query command, so no breakage there.
