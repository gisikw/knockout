---
id: ko-f65e
status: blocked
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Consolidate .ko/ top-level files into .ko/config.yaml

Currently .ko/ has several top-level files: pipeline.yml, agent.lock,
agent.pid, agent.log, prompts/. Consider merging pipeline.yml and any
other project-level config into a single .ko/config.yaml.

This would give a single place for:
- Pipeline config (currently pipeline.yml)
- Agent harness overrides
- Project-level settings (prefix, default priority, etc.)

Prereq for agent harness extraction (ko-1930) — need to know the config
surface before deciding where harness overrides live.

## Notes

**2026-02-24 08:01:44 UTC:** ko: FAIL at node 'actionable' — Plan contains 3 unresolved open questions:

Q1: Should we keep backwards compatibility with pipeline.yml?

Q2: What happens to the standalone .ko/prefix file for projects that have it?

Q3: Should prefix be writable after init, or is it set-once?

These require ticket author decision before implementation.
