All 6 tasks complete. Here's a summary:

---

## Summary

**Task 1 — `specs/ticket_triage.feature`**  
Added three new scenarios: listing tickets with triage set, excluding tickets without triage, and JSON output with the triage field.

**Task 2 — `cmd_list.go:cmdTriage`**  
Added `cmdTriage` after `cmdReady`. It follows the same structure: reorder args, resolve tickets dir, parse `--limit`/`--json` flags, load all tickets, filter to `t.Triage != ""`, sort, and output. Text format: `id [status] (pN) title — triage: <value>`. JSON format reuses `ticketJSON` (which already carries `Triage`). Build verified with `go build ./...`.

**Task 3 — `main.go` switch**  
Added `case "triage": return cmdTriage(rest)` after `case "ready"`.

**Task 4 — `main.go` help text**  
Added `triage             List tickets with a triage value set` to the commands section.

**Task 5 — `cmd_serve.go` whitelist**  
Added `"triage": true` to the whitelist map.

**Task 6 — `testdata/ticket_triage/triage_list.txtar`**  
New test file with three assertions: triage output contains triaged ticket ID and value, excludes tickets without triage, and JSON output contains the triage field. `go test ./... -run TestTicketTriage` passes.

No deviations from the plan.
