---
id: ko-bfc5
status: blocked
deps: []
links: []
created: 2026-02-24T06:11:22Z
type: task
priority: 2
---
# JSONL append on every ticket touch and build step transition

## Notes

**2026-02-24 08:49:58 UTC:** ko: FAIL at node 'actionable' — Plan has 4 open questions requiring human input:
1. Merge vs Override Semantics: Should allowed_tools at node level merge (union) with parent lists or completely replace them?
2. Tool name format: Should tool names be case-sensitive or normalized?
3. Interaction with allow_all_tool_calls: When both are set, which takes precedence?
4. Empty list semantics: Does allowed_tools: [] mean 'allow nothing' or 'inherit from parent'?
