Perfect! I've written a plan for implementing skill support in pipeline workflows. The plan:

**Goal**: Add configuration surface for skills in pipeline nodes (skills directories, skill invocation, inline skill references)

**Key points**:
- Adds `Skills []string` and `Skill string` fields to the Node struct
- Updates YAML parser to handle these new fields (following existing patterns for routes)
- Adds validation that `skill:` is mutually exclusive with `prompt:` and `run:`
- Defers actual implementation to ko-1930 (multi-agent harness) - Claude adapter will return a helpful error message
- Includes comprehensive tests following the project's testing patterns
- Documents the feature with commented examples in pipeline configs

The plan respects the ticket's explicit guidance: design the config surface now but mark Claude adapter support as pending until the multi-agent harness work (ko-1930) is complete.
