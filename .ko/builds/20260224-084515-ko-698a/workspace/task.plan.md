Perfect! I've created a comprehensive plan for implementing granular tool permissions in the pipeline config. The plan includes:

## Summary

**Goal**: Enable users to specify individual tools to auto-allow (like `Read`, `Write`, `Bash`) instead of the all-or-nothing `allow_all_tool_calls` boolean.

**Key Design Decisions**:
- Add `allowed_tools` as a string slice at all three config levels (pipeline, workflow, node)
- Use **merge semantics** (union) rather than override — node, workflow, and pipeline tool lists are combined
- Extend the harness template system with a new `${allowed_tools}` variable
- Update the `AgentAdapter` interface to pass the tools list to agent CLIs
- Claude will use `--allowed-prompts Read,Write,Bash` format

**Open Questions** that need your input:
1. **Merge vs Override**: Should node-level `allowed_tools` merge with parent lists (union) or replace them completely?
2. **Tool name format**: Are names case-sensitive? (Assumed yes, matching Claude's exact tool names)
3. **Interaction with allow_all**: Does `allow_all_tool_calls: true` override `allowed_tools`? (Assumed yes)
4. **Empty list behavior**: Does `allowed_tools: []` mean "inherit" or "allow nothing"? (Assumed inherit)

The plan breaks down the implementation into 14 concrete tasks covering parsing, resolution logic, adapter interface changes, harness templates, tests, and documentation. Each task has a verification step.
