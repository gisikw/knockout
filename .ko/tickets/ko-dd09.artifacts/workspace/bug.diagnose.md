Perfect! Now let me create a brief summary for the user:

## Diagnosis Complete

I've identified the root cause of the ticket classification confusion issue.

**The Problem**: 
The `runPromptNode` function in `build.go:466-470` unconditionally injects prior context (including previous node outputs) for ALL prompt nodes. When a ticket is built multiple times, decision nodes like `classify` receive their own previous classification output, which confuses Haiku into reclassifying tickets incorrectly.

**Root Cause**:
Decision nodes should make fresh, idempotent evaluations. By feeding them their own previous output, we create a feedback loop where the model second-guesses its prior decision instead of making a clean classification based solely on the ticket content.

**The Fix**:
Add a simple conditional check in `build.go:467` to only inject prior context for action nodes (which benefit from incremental context) but NOT for decision nodes:

```go
if node.Type == NodeAction {
    if priorContext := InjectPriorContext(artifactDir, wfName); priorContext != "" {
        prompt.WriteString(priorContext)
        prompt.WriteString("\n\n")
    }
}
```

**Affected Code**:
- Primary: `build.go:466-470` (runPromptNode function)
- Supporting: `workspace.go:36-82` (InjectPriorContext - works correctly, just needs conditional usage)

**Risk**: Very low. The change is localized, well-scoped, and all existing decision nodes (`classify`, `actionable`, `assess`, `review`) will benefit from cleaner prompts without their own output echoed back to them.

Full diagnosis written to: `.ko/tickets/ko-dd09.artifacts/diagnosis.md`
