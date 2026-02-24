---
id: ko-4895
status: open
deps: [ko-f21c]
links: []
created: 2026-02-24T07:27:01Z
type: task
priority: 2
---
# Support skills in pipeline workflows

Allow pipeline nodes to invoke skills rather than just raw prompts.

Progression:
1. `skills: [skill-dir-path]` on a node — makes skills available to the agent
2. `skill: skill-name` on a node — implies the prompt text is "apply /skill-name"
3. Inline prompt can reference skills: "apply the /whatever skill to ${TICKET_ID}"

Open question: Claude Code has no `--add-skills-dir` flag. Options:
- Symlink `.claude/commands/` to point at skill dirs (hacky)
- No-op for claude adapter, defer to multi-agent harness adapter
- Wait for upstream support

Probably best to design the config surface now but mark claude adapter
support as pending. The multi-agent harness adapter (ko-1930) will handle
this properly.
