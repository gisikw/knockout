## Summary

Implemented `ko triage` (no-args form) that lists all tickets with a non-empty `triage` field, mirroring the structure of `cmdReady`.

### What was done

1. **`cmd_list.go`** — Added `cmdTriage` function after `cmdReady`. Filters all tickets where `t.Triage != ""`, sorts by priority then modified, supports `--json`, `--limit`, and `--project` flags. Text output format: `<id> [<status>] (p<N>) <title> — triage: <value>`. Empty list returns 0 (not an error).

2. **`main.go`** — Added `case "triage": return cmdTriage(rest)` to the dispatch switch, and added `triage` to the help text alongside `ready`.

3. **`cmd_serve.go`** — Added `"triage": true` to the HTTP subcommand whitelist.

4. **`specs/ticket_triage.feature`** — Added three new scenarios: listing tickets with triage set (including across statuses), excluding tickets without triage, and JSON output.

5. **`testdata/ticket_triage/triage_list.txtar`** — New test file with three test cases matching the three spec scenarios. Uses `ko-triaged` (has triage) and `ko-plain` (no triage) tickets.

### Notable decisions

- **All statuses included**: The filter is purely `t.Triage != ""` with no status restriction, matching the ticket's "similar to ko ready" intent (ready includes open+in_progress; triage shows all triaged regardless of status since triage annotation is a human-attention signal, not a workflow state).
- **No new struct changes**: `ticketJSON` already had a `Triage` field from the `ko-13af` dependency, so JSON output required no modifications.

### Tests

All tests pass (`go test ./... -count=1`). `cmd_list.go` is 352 lines, well under the 500-line invariant.
