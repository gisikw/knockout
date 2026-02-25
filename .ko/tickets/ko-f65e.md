---
id: ko-f65e
status: open
deps: []
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

