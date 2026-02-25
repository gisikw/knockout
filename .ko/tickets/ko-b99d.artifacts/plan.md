## Goal
Remove duplicate `ko closed` and `ko resolved` commands in favor of `ko ls --status`.

## Context
The codebase has convenience commands `ko closed` and `ko resolved` that are pure aliases for `ko ls --status=closed` and `ko ls --status=resolved`. These commands exist in three locations:

1. **main.go:46-49** - Switch case entries that route to `cmdClosed` and `cmdResolved`
2. **main.go:133-134** - Help text entries listing these commands
3. **cmd_list.go:388-536** - Two complete function implementations (`cmdClosed` and `cmdResolved`) that duplicate the filtering logic already in `cmdLs`
4. **cmd_serve.go:330-331** - Whitelist entries allowing these commands via HTTP API
5. **cmd_serve_test.go:31-32, 214-215** - Test whitelist entries

The commands provide no additional functionality beyond what `ko ls --status=closed` and `ko ls --status=resolved` already offer. Both functions have identical structure to `cmdLs` with hardcoded status filters.

Per INVARIANTS.md, this is straightforward deletion - no architectural decisions needed since the functionality is preserved via `ko ls --status`.

## Approach
Delete the command implementations from cmd_list.go, remove switch cases and help text from main.go, and remove whitelist entries from cmd_serve.go and its test file. The `ko ls --status=closed` and `ko ls --status=resolved` patterns remain fully functional.

## Tasks
1. [cmd_list.go:388-536] — Delete `cmdClosed` and `cmdResolved` functions entirely. These are complete duplicates of `cmdLs` with hardcoded filters.
   Verify: `go build .` compiles successfully.

2. [main.go:46-49] — Remove `case "resolved":` and `case "closed":` switch entries that route to the deleted functions.
   Verify: `go build .` compiles successfully.

3. [main.go:133-134] — Remove help text lines for `resolved` and `closed` commands.
   Verify: `ko help` output no longer mentions these commands.

4. [cmd_serve.go:330-331] — Remove `"resolved": true,` and `"closed": true,` from the whitelist map.
   Verify: HTTP API correctly rejects `resolved` and `closed` as unallowed commands.

5. [cmd_serve_test.go:31-32, 214-215] — Remove `"resolved": true,` and `"closed": true,` from both whitelist definitions in tests, and remove `"resolved", "closed"` from the `expectedCommands` slice.
   Verify: `go test ./cmd_serve_test.go` passes.

6. [.] — Run full test suite.
   Verify: `go test ./...` passes with no failures.

## Open Questions
None. This is pure deletion of duplicate functionality. The underlying capability (`ko ls --status=closed` and `ko ls --status=resolved`) remains unchanged.
