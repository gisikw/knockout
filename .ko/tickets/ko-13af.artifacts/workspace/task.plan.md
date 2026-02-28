Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-13af.artifacts/plan.md`.

**Summary:** Adding `triage` follows the exact same pattern as `snooze`/`assignee` — it's a free-text optional string with no validation needed. The work touches 7 files across 9 ordered tasks:

1. `ticket.go` — struct field, `FormatTicket`, `ParseTicket` (3 changes)
2. `cmd_create.go` — `--triage` flag
3. `cmd_update.go` — `--triage` flag  
4. `cmd_show.go` — JSON struct + text display
5. `specs/ticket_triage.feature` — behavioral spec (required by INVARIANTS)
6. `testdata/ticket_triage/triage_basic.txtar` — integration test
7. `ticket_test.go` — parse/format unit tests

No open questions — the field is pure free-text with a clear existing pattern to follow.
