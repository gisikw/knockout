## Goal
Refactor `ko start`, `ko close`, and `ko status` to delegate status changes to `ko update` rather than implementing their own status-setting logic.

## Context

Current implementation (cmd_status.go:9-77):
- `cmdStatus()` loads ticket, validates status, sets status field, saves ticket, emits mutation event
- `cmdStart()`, `cmdClose()`, and `cmdOpen()` are thin wrappers that append a status argument and call `cmdStatus()`
- `cmdBlock()` (lines 79-108) already demonstrates the desired pattern - it wraps `cmdUpdate()` instead of implementing its own logic

Relevant files:
- cmd_status.go - contains cmdStatus, cmdStart, cmdClose, cmdOpen, cmdBlock, and ValidatePlanQuestions
- cmd_update.go - contains cmdUpdate with --status flag support (lines 218-225)
- cmd_status_test.go - contains TestValidatePlanQuestions and TestCmdBlock
- cmd_update_test.go - contains TestCmdUpdateStatus and other update tests
- main.go - command dispatch routing

The `cmdUpdate()` function already supports the `--status` flag with validation (line 219-224). It validates status using `ValidStatus()`, updates the ticket, saves it, and emits mutation events.

Key constraint: `cmdBlock()` should remain in cmd_status.go since it's a higher-level wrapper that adds logic beyond simple status setting (handling --questions flag and -d for reason). `ValidatePlanQuestions()` is only called by cmdBlock, so it belongs in the same file.

## Approach

Replace the `cmdStatus()` implementation to call `cmdUpdate()` instead of implementing its own load-validate-save logic. Update `cmdStart()`, `cmdClose()`, and `cmdOpen()` to call `cmdUpdate()` directly instead of calling `cmdStatus()`. This follows the pattern already established by `cmdBlock()`.

## Tasks

1. [cmd_status.go:9-53] — Rewrite `cmdStatus()` to delegate to `cmdUpdate()`.
   - Parse args to extract ticket ID and new status
   - Validate args count (need at least 2: ID and status)
   - Construct args for cmdUpdate: `[ticketID, "--status=<status>"]`
   - Return the exit code from cmdUpdate
   - Remove the now-unused load/validate/save/emit logic
   Verify: Existing tests in cmd_status_test.go pass (currently only tests cmdBlock and ValidatePlanQuestions, no direct cmdStatus tests)

2. [cmd_status.go:55-61] — Rewrite `cmdStart()` to call `cmdUpdate()` directly.
   - Change from calling `cmdStatus(append(args, "in_progress"))` to calling `cmdUpdate(append(args, "--status=in_progress"))`
   Verify: Manual test `ko start <id>` sets status to in_progress

3. [cmd_status.go:63-69] — Rewrite `cmdClose()` to call `cmdUpdate()` directly.
   - Change from calling `cmdStatus(append(args, "closed"))` to calling `cmdUpdate(append(args, "--status=closed"))`
   Verify: Manual test `ko close <id>` sets status to closed

4. [cmd_status.go:71-77] — Rewrite `cmdOpen()` to call `cmdUpdate()` directly.
   - Change from calling `cmdStatus(append(args, "open"))` to calling `cmdUpdate(append(args, "--status=open"))`
   Verify: Manual test `ko open <id>` sets status to open

5. [cmd_status_test.go] — Add tests for cmdStart, cmdClose, cmdOpen, and cmdStatus.
   - Add table-driven tests covering success cases and error cases (missing ID, invalid status)
   - Verify the commands properly delegate to cmdUpdate and produce correct status changes
   Verify: `go test ./... -run TestCmd(Start|Close|Open|Status)` passes

6. Run full test suite to ensure no regressions.
   Verify: `go test ./...` passes with all existing tests

## Open Questions

None. The approach is straightforward - cmdBlock already demonstrates the delegation pattern. The refactor simplifies cmd_status.go by removing duplicate status-setting logic that already exists in cmdUpdate.
