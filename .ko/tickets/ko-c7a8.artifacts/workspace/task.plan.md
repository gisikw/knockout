The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-c7a8.artifacts/plan.md`.

## Summary

I've investigated the codebase and created a plan to refactor `ko start`, `ko close`, and `ko status` to use `ko update` internally. 

**Key findings:**
- `cmdUpdate()` already supports the `--status` flag with full validation
- `cmdBlock()` already demonstrates the desired delegation pattern (wrapping `cmdUpdate`)
- The current `cmdStatus()` duplicates logic that exists in `cmdUpdate()` (load, validate, save, emit events)

**The plan involves:**
1. Rewriting `cmdStatus()` to delegate to `cmdUpdate()` 
2. Updating `cmdStart()`, `cmdClose()`, and `cmdOpen()` to call `cmdUpdate()` directly
3. Adding tests for these command wrappers
4. Running the full test suite to verify no regressions

This simplification removes ~40 lines of duplicate status-setting logic and makes the codebase more maintainable.
