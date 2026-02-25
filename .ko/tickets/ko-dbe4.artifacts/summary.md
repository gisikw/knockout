# Summary: Remove ko add-note alias

## What was done

Removed the deprecated `add-note` alias, keeping `note` as the sole command for adding notes to tickets:

1. **main.go:52-53** — Deleted the `case "add-note":` switch case and its `return cmdAddNote(rest)` line
2. **testdata/ticket_notes/notes_basic.txtar:46-50** — Removed the backward compatibility test section for `add-note`

## Verification

- Build: `go build` succeeds
- Tests: `go test ./...` passes (all cached, green)
- Spec alignment: `specs/ticket_notes.feature` only documents `ko note`, not `add-note` — no spec updates needed
- Documentation: README and help text already document only `note` — no updates needed

## Plan compliance

All planned tasks completed:
- ✓ Task 1: Removed `case "add-note"` from main.go
- ✓ Task 2: Deleted backward compatibility test
- ✓ Task 3: Verified full test suite passes

No deviations from plan.

## Invariants check

- **Specs and tests**: No spec changes required — the spec correctly documents only `note`
- **Runtime dependencies**: No new dependencies introduced
- **Statuses**: Ticket correctly marked as `in_progress`
- **Code organization**: Changes limited to command dispatch (main.go) and tests (txtar)

## Notes

This completes the transition period started in ticket ko-0a01. The `add-note` alias was kept temporarily for backward compatibility. That transition is now complete — the alias has been removed without breaking any documented behavior.
