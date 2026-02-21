---
id: ko-da10
status: closed
deps: []
links: []
created: 2026-02-20T23:45:38Z
type: task
priority: 2
---
# Per-node allow_all_tool_calls override in pipeline config

## Notes

**2026-02-20 23:45:43 UTC:** Currently allow_all_tool_calls is pipeline-wide. Add support for per-node and per-workflow overrides so triage nodes can be read-only while implement nodes get full write access. Resolution precedence should match model: node > workflow > pipeline.
