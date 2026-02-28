# Summary: ko-13af — Add triage field to ticket frontmatter

## What was done

Added a `triage` free-text string field to the `Ticket` struct and wired it end-to-end:

- **`ticket.go`**: `Triage string` field with `yaml:"triage,omitempty"`, conditionally written in `FormatTicket`, parsed in `ParseTicket`.
- **`cmd_create.go`**: `--triage` flag added to `reorderArgs` map and flag set; applied if non-empty.
- **`cmd_update.go`**: Same pattern; sets `changed = true`.
- **`cmd_show.go`**: `showJSON.Triage` field added; populated in JSON branch; displayed in text branch after `snooze`.
- **`cmd_list.go`**: `ticketJSON.Triage` field added and populated in `ticketToJSON` — not in the plan but correct and consistent with the list command's role as a full-ticket serializer.
- **`specs/ticket_triage.feature`**: Four scenarios: create with triage, update with triage, show displays triage, no triage = no field in frontmatter.
- **`testdata/ticket_triage/triage_basic.txtar`**: Integration test covering all spec scenarios.
- **`ticket_test.go`**: `TestTriageRoundTrip` and `TestParseTicketWithTriage` unit tests.
- **`ko_test.go`**: `TestTicketTriage` testscript runner.

## Notable decisions

- `cmd_list.go` was updated beyond the plan to include `Triage` in `ticketJSON`. This was the right call — `ko ls --json` is used by agents and tools that need the full ticket representation, and omitting `triage` there would be inconsistent with how all other fields are handled.

## All tests pass

`go test ./...` passes.
