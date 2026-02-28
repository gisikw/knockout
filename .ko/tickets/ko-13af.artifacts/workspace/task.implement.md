All 9 tasks complete. Here's a summary:

## Changes Made

**Task 1-3: `ticket.go`**
- Added `Triage string \`yaml:"triage,omitempty"\`` to `Ticket` struct (after `Snooze`)
- Added `if t.Triage != "" { b.WriteString(fmt.Sprintf("triage: %s\n", t.Triage)) }` in `FormatTicket` (after snooze block)
- Added `case "triage": t.Triage = val` in `ParseTicket` switch (after snooze case)

**Task 4: `cmd_create.go`**
- Added `"triage": true` to `reorderArgs` map
- Added `triage := fs.String("triage", "", "triage note (free text)")` flag
- Applied with `if *triage != "" { t.Triage = *triage }`

**Task 5: `cmd_update.go`**
- Same pattern as create: `reorderArgs` entry, flag declaration, applied with `changed = true`

**Task 6: `cmd_show.go`**
- Added `Triage string \`json:"triage,omitempty"\`` to `showJSON`
- Populated `Triage: t.Triage` in JSON branch
- Added `if t.Triage != "" { fmt.Printf("triage: %s\n", t.Triage) }` in text output

**Deviation from plan: `cmd_list.go`**
- Also added `Triage` to `ticketJSON` and `ticketToJSON` — the plan didn't explicitly mention this file, but the snooze precedent (`ko-7a57`) includes it, the integration test needed it for `ko ls --json`, and it keeps the `--json` outputs consistent.

**Task 7: `specs/ticket_triage.feature`** — New spec with 4 scenarios: create with triage, update with triage, show displays triage, ticket without triage has no triage field.

**Task 8: `testdata/ticket_triage/triage_basic.txtar`** + `ko_test.go:TestTicketTriage` — Integration test verifying create/update/show behavior.

**Task 9: `ticket_test.go`** — Added `TestTriageRoundTrip` and `TestParseTicketWithTriage` following the snooze test pattern.
