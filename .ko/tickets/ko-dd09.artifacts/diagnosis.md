# Diagnosis: Prior Output Injected into Decision Nodes Causes Classification Confusion

## Symptoms

When a ticket is built multiple times (e.g., after dependencies are resolved), the `classify` node (a decision node) receives prior context from the previous build attempt. This causes Haiku to get confused and potentially reclassify tickets incorrectly, as it sees its own previous classification output in the prompt.

### Evidence

- Build history for ko-168e shows multiple `build_start` events and duplicate `node_start` events for the classify node
- The ticket description explicitly states: "Haiku is getting confused classifying tickets that it already classified once"
- Prior output is being re-injected even though the classification decision has already been made

## Root Cause

In `build.go:466-470`, the `runPromptNode` function unconditionally injects prior context for ALL prompt nodes, regardless of whether they are decision nodes or action nodes:

```go
// Inject prior context from previous build attempts
if priorContext := InjectPriorContext(artifactDir, wfName); priorContext != "" {
    prompt.WriteString(priorContext)
    prompt.WriteString("\n\n")
}
```

The `InjectPriorContext` function in `workspace.go:36-82` collects:
1. `plan.md` from the artifact root
2. All workspace files matching the current workflow prefix (e.g., `task.*.md`, `main.*.md`)

For decision nodes like `classify`, this means:
- On the first build, the node executes cleanly and produces output (e.g., `main.classify.md`)
- On subsequent builds (retries or after unblocking), the same node receives its own previous output as "prior context"
- This confuses the model, especially Haiku, which may second-guess its previous classification or incorporate irrelevant context

**The fundamental issue**: Decision nodes make routing/classification choices that should be idempotent. Injecting their own previous output creates a feedback loop where the model sees its own reasoning and may alter its decision based on that, rather than making a fresh evaluation of the ticket.

Action nodes, by contrast, benefit from prior context because they perform work that builds incrementally (planning, implementation, etc.).

## Affected Code

### Primary Location
- **File**: `build.go`
- **Function**: `runPromptNode` (lines 431-521)
- **Specific Lines**: 466-470 (unconditional prior context injection)

### Supporting Code
- **File**: `workspace.go`
- **Function**: `InjectPriorContext` (lines 36-82)
  - This function correctly filters by workflow but doesn't distinguish between node types
  - It includes ALL workspace files for the current workflow, including prior decision node outputs

### Related Files
- **File**: `workflow.go`
- **Constants**: `NodeDecision` and `NodeAction` (lines 9-11)
  - These types are already defined and used throughout the codebase

## Recommended Fix

Add conditional logic to only inject prior context for **action nodes**, not decision nodes:

```go
// Inject prior context from previous build attempts
// Only inject for action nodes — decision nodes should make fresh evaluations
if node.Type == NodeAction {
    if priorContext := InjectPriorContext(artifactDir, wfName); priorContext != "" {
        prompt.WriteString(priorContext)
        prompt.WriteString("\n\n")
    }
}
```

### Alternative Approach (More Granular)

If some decision nodes DO need prior context in the future, add a node-level configuration option:

```yaml
- name: classify
  type: decision
  prompt: classify.md
  inject_prior_context: false  # opt-out for decision nodes
```

However, the simpler approach of excluding all decision nodes is recommended because:
1. Decision nodes should be stateless/idempotent
2. Their purpose is classification/routing, not incremental work
3. No current decision nodes in the pipeline need prior context

## Risk Assessment

### Low Risk of Fix

- **Scope**: The change is localized to a single function (`runPromptNode`)
- **Logic**: Simple conditional check using existing node type information
- **Backward Compatibility**: Action nodes continue to receive prior context as before
- **Test Coverage**: Existing tests in `workspace_test.go` verify `InjectPriorContext` behavior; new tests should verify the conditional injection

### What Could Go Wrong

1. **Decision node needs prior context for a valid reason**: Unlikely based on the semantic purpose of decision nodes, but could be discovered in edge cases. Mitigation: add per-node opt-in if needed later.

2. **Breaking existing workflows**: Current workflows (`main`, `task`, `bug`, `research`) would not be negatively affected:
   - `classify` (decision): BENEFITS from not receiving prior context
   - `actionable` (decision): BENEFITS from not receiving prior context
   - `assess` (decision): BENEFITS from not receiving prior context
   - `review` (decision): BENEFITS from not receiving prior context
   - All action nodes: unchanged behavior

3. **Test failures**: The existing `workspace_test.go` tests only verify `InjectPriorContext` function behavior, not the conditional injection logic in `runPromptNode`. New integration tests should be added.

### What It Might Affect

- **Positive Impact**: Decision nodes will make cleaner, more consistent classifications without being biased by their own previous output
- **Performance**: Slightly faster decision node execution (smaller prompts)
- **Build Retries**: Tickets that previously got stuck in reclassification loops will now proceed correctly
- **Idempotency**: Builds become more idempotent — running a build twice should produce the same classification

## Implementation Notes

1. The fix should be applied in `build.go:runPromptNode` before line 467
2. Add a comment explaining why decision nodes are excluded
3. Consider adding a test case in `build_test.go` or integration test that verifies:
   - Action nodes receive prior context
   - Decision nodes do NOT receive prior context
4. Update build history test fixtures if any expect prior context in decision nodes

---

**Summary**: The bug is a simple but impactful logic error where decision nodes receive their own previous output as context, causing classification confusion. The fix is straightforward: only inject prior context for action nodes, not decision nodes. This aligns with the semantic purpose of each node type and will prevent reclassification issues.
