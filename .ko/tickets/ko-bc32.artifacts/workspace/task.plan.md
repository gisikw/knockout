Perfect! I've written the plan to the artifact directory. The plan outlines a straightforward implementation following the established pattern of wrapper commands like `cmdStart`, `cmdClose`, and `cmdOpen`.

**Summary:**
- **Goal**: Add `ko block` as a convenience wrapper over `ko update` 
- **Pattern**: Follow the existing wrapper pattern in cmd_status.go (like `cmdStart`/`cmdClose`/`cmdOpen`)
- **Key insight**: `cmdUpdate` already supports `--status`, `-d`, and `--questions`, so `cmdBlock` just needs to parse args and transform them appropriately before delegating
- **Implementation**: 4 tasks covering the command function, tests, registration in main.go, and help text
- **No open questions**: The approach is clear and follows established patterns in the codebase

The plan is ready for review and approval.
