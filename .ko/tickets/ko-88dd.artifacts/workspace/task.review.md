So this is the ONLY place in the codebase that shells out to git. The hooks shell out to whatever the user specifies, but the core `ko` binary itself doesn't depend on git - until now.

However, let me also check the second invariant violation: **missing spec**. INVARIANTS.md lines 14-31 state:

> - **Every behavior has a spec.** Behavioral specs live in `specs/*.feature`
> - **Every spec has a test.** Go tests using `testscript` in `testdata/*.txtar`
> - **Spec before code.** Every new behavior gets a spec before or alongside the implementation.

The implementation added Go unit tests but no Gherkin spec. This is a violation.

Let me now write my review:

## Review Summary

I've completed my review of the implementation against the plan. The implementation has **two critical invariant violations** that must be addressed:

### ❌ Critical Issue 1: Runtime Dependency on Git (INVARIANTS.md Violation)

**Invariant violated:** INVARIANTS.md lines 37-41
> **Zero external runtime dependencies.** `ko` is a single static binary. It must not shell out to other tools (no `git`, `jq`, `awk`, `sed` at runtime).

**Violation location:** `build.go:807-836` - `isWorkingTreeClean()` shells out to `git status --porcelain`

**Implementation's justification:** The task.implement.md notes acknowledge this but argue the feature is "optional (like hooks)". However:
- Hooks execute **user-specified** shell commands (user's choice to depend on external tools)
- This feature makes **ko itself** depend on git being available at runtime
- This is the ONLY place in the non-test codebase that directly shells out to git

**Why this matters:** A user could have `require_clean_tree: true` in their pipeline.yml and then run `ko` in an environment without git (container, CI runner, etc.) and get cryptic failures instead of the promised "single static binary" behavior.

### ❌ Critical Issue 2: Missing Behavioral Spec (INVARIANTS.md Violation)

**Invariant violated:** INVARIANTS.md lines 14-31
> - **Every behavior has a spec.** Behavioral specs live in `specs/*.feature`
> - **Spec before code.** Every new behavior gets a spec before or alongside the implementation.

**Missing artifact:** No corresponding `specs/pipeline.feature` update or new spec file

**Current state:** The implementation added Go unit tests (TestParseRequireCleanTree, TestIsWorkingTreeClean, TestRequireCleanTreeRejectsDirty, TestRequireCleanTreeIgnoresKoDir) but no Gherkin behavioral specification.

### ✅ What Was Done Well

**Plan compliance:** All 8 tasks from the plan were completed:
1. ✅ Added `RequireCleanTree bool` field to Pipeline struct (pipeline.go:34)
2. ✅ Added YAML parsing for `require_clean_tree` (pipeline.go:221)
3. ✅ Implemented `isWorkingTreeClean()` helper (build.go:801-836)
4. ✅ Enhanced BuildEligibility to check working tree (build.go:28-68)
5. ✅ Updated RunBuild call site (build.go:61, cmd_build.go:40-68)
6. ✅ Added comprehensive Go unit tests (build_test.go:169-286)
7. ✅ Added pipeline parsing tests (pipeline_test.go:997-1039)
8. ✅ Updated example files (examples/*/pipeline.yml)

**Correctness:** The implementation matches the ticket requirements and resolved the semantic ambiguity correctly (prevent when dirty, not when clean).

**Code quality:** 
- Clean function signatures
- Proper error handling
- Good test coverage at the unit level
- Clear error messages

**Scope:** No scope creep - stayed focused on the specific feature

### 🔧 Required Fixes

I cannot fix these violations myself because:

1. **Git dependency:** The alternative approaches are:
   - Use a Go git library (go-git) - adds a dependency but keeps it in-process
   - Implement git index reading natively - very complex
   - Make this a warning instead of hard failure when git unavailable
   - Remove the feature entirely

2. **Missing spec:** I need to understand the project's spec writing conventions and add appropriate Gherkin scenarios to `specs/pipeline.feature`

Both of these require architectural decisions that go beyond simple code fixes.

```json
{
  "disposition": "fail",
  "reason": "Implementation violates two critical invariants: (1) INVARIANTS.md lines 37-41 'Zero external runtime dependencies' - isWorkingTreeClean() shells out to git at runtime, making ko depend on external git binary; (2) INVARIANTS.md lines 14-31 'Every behavior has a spec' - no Gherkin spec added to specs/pipeline.feature. The implementation is otherwise correct and complete, but these invariant violations are blockers that require architectural decisions (use go-git library vs native implementation vs degrade gracefully; spec writing conventions)."
}
```
