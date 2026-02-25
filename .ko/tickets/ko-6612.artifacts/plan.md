## Goal
Implement `ko update` command for universal ticket field mutations with auto-unblocking logic.

## Context
The codebase follows a clear pattern for ticket commands:
- Command files: `cmd_*.go` with matching test files `cmd_*_test.go`
- Each command uses `resolveProjectTicketsDir()` to handle optional `--project` flag
- Commands use `ResolveID()` for fuzzy ID matching
- Field mutations call `SaveTicket()` then `EmitMutationEvent()` for observability
- Flag parsing uses standard library `flag` package with `reorderArgs()` utility
- The `Ticket` struct (ticket.go:18-37) contains all mutable fields
- Auto-unblocking exists in `cmd_triage.go:handleAnswers()` (lines 243-245): when all `PlanQuestions` are answered, status transitions to `open`

Existing commands for reference:
- `cmd_create.go`: complex flag parsing with `-d`, `-t`, `-p`, `-a`, `--parent`, `--external-ref`, `--design`, `--acceptance`, `--tags`
- `cmd_status.go`: simple two-arg command (ID + new status) with validation
- `cmd_triage.go`: handles `--questions` (JSON) and `--answers` (JSON) with auto-unblocking

The ticket spec says tags should *replace*, not append. The spec also requires auto-unblocking: if ticket is `blocked` with questions and `--answers` resolves all questions, transition to `open`.

## Approach
Create `cmd_update.go` following the established command pattern. Parse flags for all mutable fields, apply them to the loaded ticket, handle the special case of `--answers` triggering auto-unblocking when questions are fully resolved, then save. For `--questions` and `--answers`, delegate to the existing validation and auto-unblocking logic from `cmd_triage.go` to avoid duplication.

## Tasks
1. [cmd_update.go] — Create new file with `cmdUpdate()` function following the pattern from `cmd_status.go` and `cmd_create.go`. Parse flags: `--title`, `-d` (description), `-t` (type), `-p` (priority), `-a` (assignee), `--parent`, `--external-ref`, `--design`, `--acceptance`, `--tags`, `--questions`, `--answers`, `--status`. Use `resolveProjectTicketsDir()` for project resolution, `ResolveID()` for ID matching, `flag.FlagSet` for parsing. Apply non-empty flags to ticket fields. For `--tags`, split on comma and replace the slice entirely (not append). For `--design` and `--acceptance`, append sections to body (match create behavior). For `--questions`, parse JSON, validate with `ValidatePlanQuestions()`, set `PlanQuestions` field, transition status to `blocked`. For `--answers`, parse JSON, match questions, add notes, remove answered questions from slice, auto-unblock (set status to `open`) when slice becomes empty. Call `SaveTicket()` then `EmitMutationEvent()` with event type "update". Return 0 on success.
   Verify: `go build .` compiles without errors.

2. [main.go:23-69] — Add case `"update": return cmdUpdate(rest)` in the switch statement in the `run()` function, alphabetically after `"undep"` and before `"help"`.
   Verify: `go build .` compiles without errors.

3. [main.go:110-155] — Add help text for `update` command in `cmdHelp()` function. Insert after the `triage` block and before `dep`, following the format: `update <id> [--title title] [-d description] [-t type] [-p priority] [-a assignee]\n               [--parent id] [--external-ref ref]\n               [--design notes] [--acceptance criteria]\n               [--tags tag1,tag2]\n               [--questions '<json>'] [--answers '<json>']\n               [--status status]\n                     Update ticket fields (tags replace, --answers auto-unblocks)`.
   Verify: `ko help` shows the new command in the output.

4. [cmd_update_test.go] — Create test file with minimal coverage: test updating title, description, type, priority, tags (verify replacement not append), status, and the auto-unblock behavior (blocked ticket with questions + full --answers -> status becomes open). Use `testscript` pattern if the project uses it, otherwise use standard `testing` package with temp directories like `cmd_create_test.go`. Test error cases: missing ID, invalid status, invalid question JSON.
   Verify: `go test ./... -run TestUpdate` passes.

## Open Questions
None — the behavior is fully specified in the ticket and all patterns exist in the codebase.
