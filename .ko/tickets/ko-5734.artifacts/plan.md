## Goal
Remove the redundant `ko query` command now that list commands support `--json`.

## Context
The `ko query` command outputs all tickets as JSONL (one JSON object per line). It:
- Lives in `cmd_query.go` (72 lines)
- Uses `ticketToJSON()` helper that's also used by other commands
- Is dispatched from `main.go:54-55`
- Appears in help text at `main.go:141`
- Is whitelisted in `cmd_serve.go:330`
- Has tests at `testdata/ticket_query/*.txtar` and test function at `ko_test.go:40-42`
- Has a spec at `specs/ticket_query.feature`

The `ko ls`, `ko ready`, `ko blocked` commands now all support `--json` output (as JSON arrays, not JSONL). The `ko serve` SSE endpoint uses `ListTickets()` directly via `broadcastToProject()` in `cmd_serve.go:174-203`, so removing the command won't break the HTTP API.

The `ticketToJSON()` function is shared by both `cmd_query.go` and `cmd_list.go`, so it must be preserved.

Per the ticket description: `ko ls`, `ko ready`, `ko blocked`, `ko closed`, and `ko resolved` collectively cover every status with `--json` support. Note: `ko closed` and `ko resolved` were already removed in ko-b99d, but the equivalents (`ko ls --status=closed` and `ko ls --status=resolved`) exist.

## Approach
Remove the `ko query` command entirely: delete the source file, remove the dispatcher case, update help text, remove the serve whitelist entry, delete tests and specs. Keep `ticketToJSON()` by moving it to `cmd_list.go` since that's where it's primarily used now.

## Tasks
1. [cmd_query.go] — Delete the entire file.
   Verify: File no longer exists.

2. [cmd_list.go:10-49] — Move `ticketToJSON()` function from `cmd_query.go` to the top of `cmd_list.go` (it's already imported there, so no changes needed to imports).
   Verify: `cmd_list.go` compiles and the function is accessible where needed.

3. [main.go:54-55] — Remove the `case "query":` switch entry and its `return cmdQuery(rest)` line.
   Verify: `go build` succeeds, `ko query` returns "unknown command" error.

4. [main.go:141] — Remove the `query                 Output all tickets as JSONL` line from help text.
   Verify: `ko help` output doesn't mention query.

5. [cmd_serve.go:330] — Remove `"query": true,` from the whitelist map.
   Verify: HTTP API rejects query commands with "not allowed" error.

6. [ko_test.go:40-42] — Delete the `TestTicketQuery` function.
   Verify: `go test` doesn't attempt to run query tests.

7. [testdata/ticket_query/] — Delete the entire directory and its contents.
   Verify: Directory no longer exists, tests still pass.

8. [specs/ticket_query.feature] — Delete the spec file.
   Verify: File no longer exists.

9. Run `go test ./...` to confirm all tests pass without the query command.
   Verify: All tests pass with no failures.

## Open Questions
None. The ticket is explicit: `ko query` is redundant now that list commands support `--json`. The SSE endpoint doesn't use the command (it uses `ListTickets()` directly), so no breakage there. The `ticketToJSON()` helper must be preserved for the other commands that still need it.
