## Goal
Add a `ko block` convenience wrapper that sets status to blocked with an optional reason or questions.

## Context
The `ko update` command (cmd_update.go) already supports:
- `--status blocked` to set ticket status
- `-d <description>` to append to the body
- `--questions <json>` to add questions (automatically sets status to blocked)

The ticket wants two forms:
1. `ko block <id> [reason]` → maps to `ko update <id> --status=blocked -d reason`
2. `ko block <id> --questions json` → maps to `ko update <id> --status=blocked --questions json`

Existing wrapper pattern (cmd_status.go:55-77):
- `cmdStart`, `cmdClose`, `cmdOpen` are thin wrappers over `cmdStatus`
- They append the target status to args and delegate to `cmdStatus`
- Each validates that a ticket ID was provided before delegating

The `ko block` command should follow this pattern but delegate to `cmdUpdate` instead, since it needs to handle both `--questions` (already supported by `cmdUpdate`) and an optional reason text.

Command registration happens in main.go switch statement (line 23-71).

Tests exist in cmd_status_test.go for the wrapper pattern, and cmd_update_test.go for the underlying functionality.

## Approach
Add a `cmdBlock` function in cmd_status.go (where the other status wrapper commands live). Parse the args to detect whether `--questions` is present. If questions are provided, pass through to `cmdUpdate` with `--questions`. If a positional reason is provided, transform it to `--status=blocked -d reason`. The command should reuse `cmdUpdate`'s existing validation and logic. Register the command in main.go switch. Add Go tests in cmd_status_test.go following the existing wrapper pattern.

## Tasks
1. [cmd_status.go] — Add `cmdBlock(args []string) int` function after `cmdOpen` (line 78).
   - Parse args to check for `--questions` flag
   - If `--questions` present: delegate to `cmdUpdate` with original args plus `--status=blocked`
   - If positional reason provided: delegate to `cmdUpdate` with `<id>`, `--status=blocked`, `-d`, `<reason>`
   - If no reason/questions: delegate to `cmdUpdate` with `<id>`, `--status=blocked`
   - Validate ticket ID presence before delegation
   Verify: `go test ./... -run TestCmdBlock` passes

2. [cmd_status_test.go] — Add `TestCmdBlock` function at end of file.
   - Test case: block with reason → status set to blocked, body contains reason
   - Test case: block with --questions → status set to blocked, questions added
   - Test case: block without reason → status set to blocked only
   - Test case: missing ticket ID → error
   Verify: `go test ./... -run TestCmdBlock` passes

3. [main.go:46] — Add case for "block" command after "blocked" case.
   - `case "block": return cmdBlock(rest)`
   Verify: `ko block` command recognized (run `ko block` and see usage error, not "unknown command")

4. [main.go:122] — Update help text to document the block command.
   - Add line in command list: `block <id> [reason]    Block ticket with optional reason`
   - Add `block <id> --questions '<json>'  Block ticket with questions`
   Verify: `ko help` shows the new command

## Open Questions
None — the implementation is straightforward wrapper following the established pattern. The `cmdUpdate` command already supports all the underlying functionality (`--status`, `-d`, `--questions`), so `cmdBlock` just needs to transform args and delegate.
