---
id: ko-1930
status: open
deps: []
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Extract agent harnesses into file-parameterized configs

Agent runners (claude, cursor) are currently hardcoded in Go (adapter.go).
Extract them into declarative config files so they're extensible without
recompiling.

Proposed locations:
- `./agent-harnesses/{claude,cursor}.yaml` — shipped with ko
- `~/.config/knockout/agent-harnesses/` — user-supplied supplemental harnesses

Harness config needs to describe:
- How to build the command (binary, args template)
- How to pass prompt, model, system prompt, allow_all
- How to pass env vars (workspace, artifact dir, etc.)
- Format could be yaml, sh, or a hybrid — whatever makes the template clean

The `agent:` field in pipeline.yml would then resolve against both built-in
and user-supplied harnesses.

Consider prereq: ko-f65e (consolidate .ko/ config) — where do harness
overrides live in the project config?
