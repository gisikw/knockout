---
id: ko-dd51
status: open
deps: []
created: 2026-02-24T12:47:30Z
type: task
priority: 2
---
# New ticket status: resolved — done but pending human review before close

Add a `resolved` status between `in_progress` and `closed`. For work
that's been completed but needs human sign-off before closing.

Use cases:
- Research questions where the agent found an answer but a human should
  confirm the conclusion before closing
- Bug wontfixes where the decision to not fix should be reviewed
- Any agent-completed work that warrants a sanity check

`ko ready` should not surface resolved tickets. `ko resolved` (or
`ko ls --status resolved`) should list them. The agent loop should
transition tickets to `resolved` instead of `closed` for research
and bug-wontfix workflows.
