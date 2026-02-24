---
id: ko-fd4e
status: open
deps: []
created: 2026-02-24T12:47:30Z
type: task
priority: 2
---
# Add .ko/.gitignore for agent runtime files (agent.lock, agent.pid, agent.log)

Write a `.ko/.gitignore` covering runtime state files that shouldn't be
tracked: agent.lock, agent.pid, agent.log. Should be created by
`ko agent init`. See ko-1390 research findings for context.

## Notes

**2026-02-24 13:06:57 UTC:** ko: FAIL at node 'classify' — node 'classify' failed after 3 attempts: disposition missing required 'disposition' field
