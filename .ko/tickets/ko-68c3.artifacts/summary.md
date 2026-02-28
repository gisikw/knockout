## Summary

Implemented `ko triage <id> <instructions>` as an alias for `ko update <id> --triage=<instructions>`.

### What was done

1. **`cmd_list.go`**: Added a branch in `cmdTriage` after flag parsing. If two or more positional args are present, delegates to `cmdUpdate` with `--triage=<joined instructions>`. If exactly one positional arg is present, prints an error to stderr and returns 1. Zero args preserves the existing list behavior.

2. **`main.go`**: Updated the help text for `triage` to document both forms: `triage [<id> <instructions>]   List tickets with a triage value, or set triage on a ticket`.

3. **`cmd_list_test.go`**: Added `TestCmdTriageSet` covering: setting triage saves the field, multi-word instructions are joined correctly, missing instructions (1 arg) returns non-zero, and 0 args still performs the list.

### Reviewer fix

The automated implementation stage did not update the spec (`specs/ticket_triage.feature`) or add a testscript test (`testdata/ticket_triage/triage_set.txtar`), both required by INVARIANTS.md ("every behavior has a spec", "every spec has a test"). These were added during review:

- `specs/ticket_triage.feature`: Three new scenarios for set, multi-word join, and missing-instructions error.
- `testdata/ticket_triage/triage_set.txtar`: Corresponding testscript tests.

All tests pass (`go test ./...`).

### Notable decisions

- Follows the same delegation pattern as `cmdSnooze` (delegates to `cmdUpdate` with a constructed flag) and `cmdBlock` (joins args for multi-word values).
- No new logic was needed beyond the branch — `cmdUpdate` already handles `--triage=`.
