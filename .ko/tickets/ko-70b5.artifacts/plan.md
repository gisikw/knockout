## Goal
Add `ko triage` with no args as a command that lists all tickets with a triage value set, mirroring the structure of `ko ready`.

## Context
- `cmdReady` lives in `cmd_list.go:200`. It loads all tickets, filters with `IsReady + !IsSnoozed + t.Triage == ""`, and outputs them. The `triage` command inverts that last condition.
- `ticketJSON` already has `Triage string` (line 29 of `cmd_list.go`), so JSON output needs no new struct changes.
- `main.go` dispatches commands via a switch; adding `"triage"` there is the entry point.
- `cmd_serve.go` has a whitelist map of allowed HTTP-served subcommands; `triage` must be added.
- `specs/ticket_triage.feature` exists but has no scenario for `ko triage` listing — needs new scenarios.
- `testdata/ticket_triage/` already has two `.txtar` files; a new one for the listing behavior goes there.
- `ko_test.go` already registers `TestTicketTriage` pointing at `testdata/ticket_triage/` — no new test function needed.
- `cmd_list.go` is 290 lines, well under the 500-line limit.
- The `triage` command should filter on any ticket where `t.Triage != ""`, across all statuses (open, in_progress, etc.), just as `ko ready` includes open and in_progress. The ticket doesn't restrict by status; being analogous to `ready` means listing tickets that need human triage attention regardless of their workflow state.
- Should support `--json` and `--limit` flags and `--project` / `#tag` routing, consistent with `ready`.

## Approach
Add `cmdTriage` to `cmd_list.go` following the same pattern as `cmdReady`: resolve tickets dir, load all tickets, filter to `t.Triage != ""`, sort by priority then modified, output text or JSON. Register it in `main.go`'s switch, add to `cmd_serve.go` whitelist, add to help text, add spec scenarios, and add a `.txtar` test.

## Tasks
1. **[specs/ticket_triage.feature]** — Add two new scenarios: "Triage with no args lists tickets with triage set" and "Triage with no args excludes tickets without triage set". Also add a JSON scenario for completeness.
   Verify: spec is readable and consistent with existing style.

2. **[cmd_list.go:cmdTriage]** — Add `func cmdTriage(args []string) int` after `cmdReady`. It should:
   - Call `reorderArgs(args, map[string]bool{"project": true, "limit": true})`
   - Resolve ticketsDir via `resolveProjectTicketsDir`
   - Parse `--limit` and `--json` flags (same as `cmdReady`)
   - Load all tickets with `ListTickets`
   - Filter: `t.Triage != ""`
   - Sort with `SortByPriorityThenModified`
   - Output in text format (`id [status] (pN) title — triage: <value>`) or JSON
   - Return 0 (empty list is not an error, same as `ready`)
   Verify: `go build ./...` succeeds.

3. **[main.go:run]** — Add `case "triage": return cmdTriage(rest)` to the switch.
   Verify: `ko triage` is reachable.

4. **[main.go:cmdHelp]** — Add `triage` to the help text alongside `ready`.
   Verify: `ko help` output contains "triage".

5. **[cmd_serve.go:whitelist]** — Add `"triage": true` to the whitelist map.
   Verify: HTTP server accepts `{"argv":["triage"]}` requests.

6. **[testdata/ticket_triage/triage_list.txtar]** — New test file covering:
   - `ko triage` lists tickets with triage set (shows triage value in output)
   - `ko triage` excludes tickets without triage
   - `ko triage --json` outputs valid JSON with triage field
   Verify: `go test ./... -run TestTicketTriage` passes.

## Open Questions
None. The ticket says "similar to ko ready" — the pattern is fully defined by the existing `cmdReady` implementation. Status filtering is not mentioned, so list all tickets with a non-empty triage field regardless of status (just as `ready` looks at open+in_progress rather than filtering on a single status).
