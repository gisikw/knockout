## Goal
Remove the `ko add-note` alias while keeping `ko note` as the primary command.

## Context
The codebase currently defines two identical aliases for the same functionality:
- `main.go:50-53` has a switch statement with two cases (`note` and `add-note`) both calling `cmdAddNote()`
- `cmd_note.go` implements the shared `cmdAddNote()` function
- `testdata/ticket_notes/notes_basic.txtar:46-50` includes a backward compatibility test for `add-note`
- `main.go:143` documents only `note` in the help text (not `add-note`)
- `README.md:32` documents only `note` in the usage section

The `add-note` alias was kept for backward compatibility during a transition period (see ticket ko-0a01). That transition is complete and the alias should now be removed.

## Approach
Remove the `add-note` case from the main switch statement and delete the backward compatibility test. The help text and README already document only `note`, so no documentation updates are needed.

## Tasks
1. [main.go:52-53] — Delete the `case "add-note":` line and its associated `return cmdAddNote(rest)` line.
   Verify: `go build` succeeds.

2. [testdata/ticket_notes/notes_basic.txtar:46-50] — Delete the backward compatibility test section (lines 46-50: comment + exec + stdout).
   Verify: `go test ./... -run TestScript` passes.

3. Run full test suite to confirm no other tests depend on the `add-note` alias.
   Verify: `go test ./...` passes with all tests green.

## Open Questions
None. This is a straightforward removal of a deprecated alias with clear scope.
