Perfect! All tasks have been completed successfully. Let me provide a summary of the changes.

## Summary

I successfully implemented the ticket to remove the `ko add-note` alias while keeping `ko note`. All changes were made according to the plan, with no deviations.

### Changes by task:

**Task 1: Remove `add-note` case from main.go**
- Deleted lines 52-53 in `main.go` (the `case "add-note":` and its `return cmdAddNote(rest)` statement)
- Verified: `go build` succeeded

**Task 2: Remove backward compatibility test**
- Deleted lines 46-50 in `testdata/ticket_notes/notes_basic.txtar` (the comment and test for `add-note` backward compatibility)
- Verified: `TestTicketNotes/notes_basic` test passed

**Task 3: Run full test suite**
- Ran `go test ./...` - all tests passed (6.134s)
- Confirmed no other tests depend on the `add-note` alias

The `ko note` command remains fully functional. The `add-note` alias has been completely removed from both the implementation and the tests. No documentation changes were needed as the help text and README already documented only `ko note`.
