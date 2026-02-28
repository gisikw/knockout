Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-70b5.artifacts/plan.md`.

Here's the summary:

**Goal:** Add `ko triage` (no args) to list all tickets with a `triage` field set — mirroring `ko ready`.

**Key findings:**
- `cmdReady` in `cmd_list.go:200` is the direct template. The new `cmdTriage` simply inverts the filter condition: `t.Triage != ""` instead of `t.Triage == ""`.
- `ticketJSON` already carries the `Triage` field — no struct changes needed.
- The spec file at `specs/ticket_triage.feature` exists but lacks listing scenarios.
- `ko_test.go` already runs `TestTicketTriage` against `testdata/ticket_triage/` — just add a new `.txtar`.

**Tasks (in order):**
1. Add spec scenarios to `specs/ticket_triage.feature`
2. Add `cmdTriage` to `cmd_list.go` (supports `--json`, `--limit`, `--project`/`#tag`)
3. Register `"triage"` in `main.go`'s switch
4. Add to help text in `main.go`
5. Add to the HTTP server whitelist in `cmd_serve.go`
6. Add `testdata/ticket_triage/triage_list.txtar` with text + JSON coverage
