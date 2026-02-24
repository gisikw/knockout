## Goal
Enable users to specify granular tool permissions per-node instead of all-or-nothing via `allow_all_tool_calls`.

## Context
The current pipeline configuration supports `allow_all_tool_calls` as a boolean flag at three levels (pipeline, workflow, node) with proper inheritance (node > workflow > pipeline). This maps to agent-specific flags like `--dangerously-skip-permissions` for Claude or `--force` for Cursor via the harness template system.

Key files:
- **pipeline.go**: Defines `Pipeline`, `Workflow` types; parses YAML config including `allow_all_tool_calls`
- **workflow.go**: Defines `Node` type with `AllowAll *bool` field
- **build.go**: Contains `resolveAllowAll()` to resolve the three-level override hierarchy; calls `adapter.BuildCommand()` with the resolved `allowAll` boolean
- **harness.go**: `TemplateAdapter.BuildCommand()` renders `${allow_all}` and `${cursor_allow_all}` template variables based on the boolean flag
- **adapter.go**: Defines `AgentAdapter` interface with `BuildCommand(prompt, model, systemPrompt string, allowAll bool)` signature
- **agent-harnesses/claude.yaml**: Uses `${allow_all}` which expands to `--dangerously-skip-permissions` when true
- **agent-harnesses/cursor.yaml**: Uses `${cursor_allow_all}` which expands to `--force` when true

The current system is binary: either skip all permission checks or require prompts for everything. The ticket requests granular control to specify which tools are auto-allowed.

## Approach
Add an `allowed_tools` field (string slice) to the pipeline config hierarchy (pipeline, workflow, node levels) alongside the existing `allow_all_tool_calls` boolean. Update the harness template system to support a new `${allowed_tools}` variable that expands to the appropriate CLI flag syntax for each agent. Maintain the existing three-level inheritance pattern but for string slices rather than booleans. The `AgentAdapter` interface will need to accept the tools list and agents will format it appropriately (Claude uses `--allowed-prompts`, Cursor may ignore it if unsupported).

## Tasks
1. **[pipeline.go:Pipeline]** — Add `AllowedTools []string` field to the `Pipeline` struct alongside `AllowAll`. Parse `allowed_tools:` from YAML as a string list (support both multiline and inline `[a, b]` syntax, mirroring the existing `routes`/`skills` parsing pattern). No verification needed yet.

2. **[workflow.go:Workflow,Node]** — Add `AllowedTools []string` field to both `Workflow` and `Node` structs to enable per-workflow and per-node overrides. No validation changes needed since tools lists are permissive.

3. **[pipeline.go:ParsePipeline]** — Update the YAML parser to handle `allowed_tools:` at all three levels (pipeline, workflow, node) using the same pattern as the existing `skills` parsing (lines 307-313 show the pattern: handle both inline `[a, b]` syntax and multiline list, toggle a flag to track when we're accumulating a multiline list). Verify: Add unit test in pipeline_test.go.

4. **[build.go:resolveAllowedTools]** — Create new function `resolveAllowedTools(p *Pipeline, wf *Workflow, node *Node) []string` that implements merge semantics: node tools + workflow tools + pipeline tools (union, deduplicated). This differs from `resolveAllowAll` which does override semantics. Verify: Add unit test.

5. **[build.go:runPromptNode]** — Update the call site around line 168 to resolve allowed tools and pass them to `adapter.BuildCommand()`. Change from `allowAll := resolveAllowAll(p, wf, node)` to also compute `allowedTools := resolveAllowedTools(p, wf, node)`, then pass both to the adapter. Verify: Existing build tests should still pass.

6. **[adapter.go:AgentAdapter]** — Update the `BuildCommand` interface signature from `BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd` to `BuildCommand(prompt, model, systemPrompt string, allowAll bool, allowedTools []string) *exec.Cmd`. All implementations must be updated. Verify: Compilation succeeds.

7. **[adapter.go:RawCommandAdapter.BuildCommand]** — Update method signature to accept `allowedTools []string` parameter. Do not implement any new logic (raw commands don't support this feature). Verify: Compilation succeeds.

8. **[harness.go:TemplateAdapter.BuildCommand]** — Update method signature and add `allowed_tools` to the template variable map. Format the tools list as a comma-separated value for the `${allowed_tools}` variable (e.g., `"Read,Write,Bash"`). When the list is non-empty, expand `${allowed_tools}` to `--allowed-prompts\n<comma-separated-list>`; when empty, expand to empty string (following the existing pattern for `model` and `system_prompt` at lines 86-91). Verify: Add unit test.

9. **[agent-harnesses/claude.yaml]** — Add `${allowed_tools}` to the args list (insert after `${allow_all}`, before `${model}`). This will expand to `--allowed-prompts Read,Write,Bash` when tools are specified. Verify: Manual smoke test with a real build.

10. **[agent-harnesses/cursor.yaml]** — Add `${allowed_tools}` to args for future compatibility, though Cursor may not support this flag yet. When unsupported, the agent will ignore the flag. Verify: No-op change, just forward compatibility.

11. **[pipeline_test.go]** — Add test `TestParsePipelineAllowedTools` that validates parsing of `allowed_tools` at all three levels (pipeline, workflow, node) in both inline and multiline formats. Verify: Test passes.

12. **[pipeline_test.go]** — Add test `TestResolveAllowedToolsMerge` that validates the union merge behavior across the three levels (node + workflow + pipeline with deduplication). Verify: Test passes.

13. **[README.md:Pipeline options]** — Add row to the pipeline options table: `allowed_tools` | `[]` | List of tool names to auto-allow (e.g., `Read`, `Write`, `Bash`). Can be set at pipeline, workflow, or node level; lists are merged (union). Verify: Documentation is clear.

14. **[README.md:Node properties]** — Document that `allowed_tools` can be set per-node as a list (multiline or inline syntax). Verify: Documentation is clear.

## Open Questions
1. **Merge vs Override Semantics**: Should `allowed_tools` at node level completely replace the parent lists, or should it merge (union) with workflow and pipeline lists? The plan assumes **merge semantics** (union) because this is more permissive and aligns with the principle of least surprise — a node shouldn't become more restrictive by accident. However, if users need to restrict tools at a lower level, we could support both a positive list (`allowed_tools`) and a negative list (`denied_tools`), or use override semantics. **Decision needed**: Merge or override?

2. **Tool name format**: Are tool names case-sensitive? Should we normalize them (e.g., `"read"` vs `"Read"` vs `"READ"`)? The Claude CLI uses exact matches like `"Read"`, `"Write"`, `"Bash"`. **Assumption**: Use exact case-sensitive matching as provided by the user, document the canonical names in README.

3. **Interaction with allow_all_tool_calls**: When both `allow_all_tool_calls: true` and `allowed_tools: [Read, Write]` are set, which takes precedence? **Assumption**: `allow_all_tool_calls: true` overrides everything (skips all permission checks). The `allowed_tools` list is only used when `allow_all_tool_calls` is false. Document this behavior.

4. **Empty list semantics**: Does `allowed_tools: []` mean "allow nothing" (maximally restrictive) or "inherit from parent"? **Assumption**: Empty list means inherit from parent (no tools specified at this level). To block all tools, omit the `allowed_tools` field entirely and set `allow_all_tool_calls: false`. Document this.
