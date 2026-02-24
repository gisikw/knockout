# Findings: allowed_tools Configuration Semantics

## Summary

Investigated the four open questions about `allowed_tools` configuration semantics by examining the existing codebase, related ticket ko-698a, and the current implementation of `allow_all_tool_calls`. One question has already been decided, and the other three have proposed assumptions that require human approval.

## Question 1: Merge vs Override Semantics ✅ RESOLVED

**Question:** Should `allowed_tools` at node level merge (union) with parent lists or completely replace them?

**Answer:** **OVERRIDE semantics** — Node-level `allowed_tools` completely replaces parent lists.

**Evidence:** Decision documented in `.ko/tickets/ko-698a.md` note from 2026-02-24 13:02:23 UTC:
> "Decision: override, not merge. Node-level allowed_tools completely replaces parent lists."

**Implementation Reference:** This differs from the existing `allow_all_tool_calls` boolean field, which also uses override semantics. See `build.go:542-552`:

```go
// resolveAllowAll returns the most specific allow_all_tool_calls override.
// Precedence: node > workflow > pipeline.
func resolveAllowAll(p *Pipeline, wf *Workflow, node *Node) bool {
	if node.AllowAll != nil {
		return *node.AllowAll
	}
	if wf.AllowAll != nil {
		return *wf.AllowAll
	}
	return p.AllowAll
}
```

The new `resolveAllowedTools()` function should follow the same pattern but for `[]string` instead of `bool`.

---

## Question 2: Tool Name Format

**Question:** Should tool names be case-sensitive or normalized?

**Current Assumption (from ko-698a plan):** Use exact case-sensitive matching as provided by the user. Document the canonical names (e.g., `Read`, `Write`, `Bash`) in README.

**Evidence from Codebase:**
- Claude CLI uses exact matches like `"Read"`, `"Write"`, `"Bash"` (referenced in ko-698a plan line 53)
- No existing normalization logic in the codebase
- The harness template system (see `harness.go:71-141`) does simple string replacement without any case transformation

**Recommendation:** **Accept the assumption** — use case-sensitive matching. This is simpler, avoids transformation bugs, and matches how the underlying agents work.

**Rationale:**
- Simplicity: No need for normalization logic
- Predictability: What you write is what you get
- Consistency: Matches the underlying Claude CLI behavior
- Documentation: Just document the canonical names in README (already planned in task #13)

---

## Question 3: Interaction with allow_all_tool_calls

**Question:** When both `allow_all_tool_calls: true` and `allowed_tools: [Read, Write]` are set, which takes precedence?

**Current Assumption (from ko-698a plan):** `allow_all_tool_calls: true` overrides everything (skips all permission checks). The `allowed_tools` list is only used when `allow_all_tool_calls` is false.

**Evidence from Codebase:**

1. **Current agent adapters:** See `adapter.go:12` and `harness.go:71`:
   ```go
   BuildCommand(prompt, model, systemPrompt string, allowAll bool) *exec.Cmd
   ```
   The `allowAll` boolean maps to `--dangerously-skip-permissions` (Claude) or `--force` (Cursor).

2. **Flag semantics:** When `--dangerously-skip-permissions` is set, the Claude CLI skips **all** permission checks. This is a complete bypass.

3. **Template expansion:** See `harness.go:92-95`:
   ```go
   if allowAll {
       vars["allow_all"] = "--dangerously-skip-permissions"
       vars["cursor_allow_all"] = "--force"
   }
   ```

**Recommendation:** **Accept the assumption** — `allow_all_tool_calls: true` takes precedence.

**Rationale:**
- Semantic clarity: "allow all" literally means "allow all tools"
- Backward compatibility: Existing configs with `allow_all_tool_calls: true` should continue to work identically
- Simplicity: Avoids complex precedence rules
- Security: Explicit trumps implicit — if someone sets `allow_all: true`, they're making an explicit choice to bypass all permissions

**Implementation Note:** The resolution logic should check `allowAll` first. If true, pass the `--dangerously-skip-permissions` flag and ignore `allowed_tools`. If false, format `allowed_tools` into `--allowed-prompts <list>`.

---

## Question 4: Empty List Semantics

**Question:** Does `allowed_tools: []` mean 'allow nothing' or 'inherit from parent'?

**Current Assumption (from ko-698a plan):** Empty list means inherit from parent (no tools specified at this level). To block all tools, omit the `allowed_tools` field entirely and set `allow_all_tool_calls: false`.

**Analysis:**

This is a **critical design decision** with significant UX implications.

### Option A: Empty List = Inherit from Parent (Current Assumption)

**Pros:**
- Allows incremental specification — you only set `allowed_tools` where you need to override
- Consistent with the "nil means inherit" pattern used for `AllowAll *bool` fields
- Simpler mental model: "If I don't specify tools at this level, use parent's setting"

**Cons:**
- Cannot express "no tools allowed at this level" without introducing a separate mechanism
- Ambiguity: `allowed_tools: []` and omitting the field mean the same thing

### Option B: Empty List = Allow Nothing (Maximally Restrictive)

**Pros:**
- Clear distinction: omit field = inherit, empty list = block all
- Enables lockdown: a node can explicitly restrict tools even if parent is permissive
- Matches common YAML convention where `[]` is a meaningful empty value

**Cons:**
- Easy to accidentally lock down by writing `allowed_tools: []` intending to inherit
- Requires all three levels (pipeline, workflow, node) to specify tools if any level has `[]`
- More complex to reason about

**Current Implementation Pattern:**

Looking at how `AllowAll *bool` works (see `workflow.go:20` and `pipeline.go:42`):
- Pipeline: `AllowAll bool` (defaults to `false`)
- Workflow: `AllowAll *bool` (nil = inherit from pipeline)
- Node: `AllowAll *bool` (nil = inherit from workflow)

This is a **nil-means-inherit** pattern for pointer types.

For `allowed_tools`, we'd use `[]string`:
- Cannot use `nil` vs `[]` distinction in the YAML (both parse to nil/empty)
- Need a different approach

**Recommendation:** **Accept the assumption** (empty list = inherit), BUT use a more explicit implementation.

**Proposed Implementation:**
1. Use `*[]string` (pointer to slice) instead of `[]string`:
   - `nil` = field not set, inherit from parent
   - `&[]string{}` = empty list explicitly set, inherit from parent
   - `&[]string{"Read", "Write"}` = tools explicitly set
2. This matches the existing pattern for `AllowAll *bool`
3. In YAML parsing, only populate the field if `allowed_tools:` is present
4. Document clearly: "Omit `allowed_tools` to inherit. Set it (even to empty list) to override parent."

**Alternative (Simpler):** Just use `[]string` and treat empty/nil as "inherit", and document that there's no way to explicitly block all tools via `allowed_tools` (you must use `allow_all_tool_calls: false` and omit `allowed_tools` at all levels).

**Decision Required:** Choose between:
- **A (Simpler):** Use `[]string`, empty = inherit, no way to explicitly block via `allowed_tools`
- **B (More Flexible):** Use `*[]string`, nil = inherit, empty = inherit, non-empty = override

---

## Current State

### Already Implemented
- Three-level configuration hierarchy for `allow_all_tool_calls` (pipeline → workflow → node)
- Override semantics with `*bool` for optional fields (nil = inherit)
- Template-based harness system that can expand new variables
- YAML parser supporting both inline `[a, b]` and multiline list syntax

### Needs Implementation (from ko-698a plan)
1. Add `AllowedTools` fields to `Pipeline`, `Workflow`, `Node` structs
2. Update YAML parser to handle `allowed_tools:` at all levels
3. Create `resolveAllowedTools()` function with override semantics
4. Update `AgentAdapter.BuildCommand()` signature to accept `allowedTools []string`
5. Update harness template system to support `${allowed_tools}` variable
6. Update agent harness configs to use the new variable
7. Add tests and documentation

### Files to Modify
- `pipeline.go` — add fields, update parser
- `workflow.go` — add fields to Node
- `build.go` — add resolver function, update call sites
- `adapter.go` — update interface signature
- `harness.go` — update template expansion
- `agent-harnesses/claude.yaml` — add `${allowed_tools}` variable
- `agent-harnesses/cursor.yaml` — add for future compatibility
- `pipeline_test.go` — add tests
- `README.md` — document the feature

---

## Recommended Actions

### For Question #1: ✅ No Action Needed
Already decided: use override semantics.

### For Question #2: Accept Assumption
**Approve:** Use case-sensitive tool names. Document canonical names in README.

### For Question #3: Accept Assumption
**Approve:** `allow_all_tool_calls: true` takes precedence over `allowed_tools`.

### For Question #4: Decision Required
**Choose one:**
- **Option A (Recommended):** Simple `[]string`, empty = inherit, document that you can't explicitly block all tools via `allowed_tools` alone
- **Option B:** More complex `*[]string` pattern for maximum flexibility

Once these decisions are made, proceed with implementation per the ko-698a plan.

---

## Code References

**Key Files:**
- `pipeline.go:11-27` — Pipeline struct definition
- `workflow.go:14-26` — Node struct definition
- `build.go:542-552` — resolveAllowAll() pattern to follow
- `adapter.go:8-13` — AgentAdapter interface (needs signature update)
- `harness.go:71-141` — TemplateAdapter.BuildCommand() template expansion
- `agent-harnesses/claude.yaml:8` — where `${allow_all}` is used
- `README.md:245-256` — Pipeline options documentation

**Related Ticket:**
- `.ko/tickets/ko-698a.md` — Parent ticket with full implementation plan
