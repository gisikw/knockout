## Goal
Add a `triage` free-text string field to ticket frontmatter, settable via `ko add --triage` and `ko update --triage`.

## Context
The codebase follows a consistent pattern for optional string fields on `Ticket` (e.g. `assignee`, `external-ref`, `snooze`). Each requires:
1. A struct field in `ticket.go` (`Ticket` struct, `FormatTicket`, `ParseTicket`)
2. A `--<field>` flag in `cmd_create.go` and `cmd_update.go` (including `reorderArgs` map entry)
3. Display in `cmd_show.go` (both text and JSON output via `showJSON`)
4. A spec in `specs/`
5. A txtar test in `testdata/`
6. Unit tests in `ticket_test.go`

`triage` is simpler than `snooze` — it's pure free-text with no validation needed. The snooze commit (`ko-7a57`) is the closest precedent.

Key files:
- `ticket.go`: `Ticket` struct, `FormatTicket`, `ParseTicket`
- `cmd_create.go`: flag declaration, `reorderArgs` map, field application
- `cmd_update.go`: same pattern
- `cmd_show.go`: `showJSON` struct + text output block
- `ticket_test.go`: parse/format unit tests
- `specs/ticket_triage.feature`: behavioral spec
- `testdata/ticket_triage/triage_basic.txtar`: integration test

## Approach
Add `Triage string` with `yaml:"triage,omitempty"` to the `Ticket` struct, wire it through `FormatTicket` and `ParseTicket`, expose it as `--triage` in both `cmd_create.go` and `cmd_update.go`, and surface it in `cmd_show.go`. Then write the spec and tests.

## Tasks

1. **[ticket.go:Ticket]** — Add `Triage string \`yaml:"triage,omitempty"\`` field to the `Ticket` struct, after `Snooze`.
   Verify: `go build ./...` passes.

2. **[ticket.go:FormatTicket]** — Add conditional output `if t.Triage != "" { b.WriteString(fmt.Sprintf("triage: %s\n", t.Triage)) }` after the `snooze` block.
   Verify: `go build ./...` passes.

3. **[ticket.go:ParseTicket]** — Add `case "triage": t.Triage = val` to the switch in the standard frontmatter parsing block (after the `snooze` case).
   Verify: `go build ./...` passes.

4. **[cmd_create.go:cmdCreate]** — Add `"triage": true` to the `reorderArgs` map; add `triage := fs.String("triage", "", "triage note (free text)")` flag; apply with `if *triage != "" { t.Triage = *triage }`.
   Verify: `go build ./...` passes.

5. **[cmd_update.go:cmdUpdate]** — Add `"triage": true` to the `reorderArgs` map; add `triage := fs.String("triage", "", "triage note (free text)")` flag; apply with `if *triage != "" { t.Triage = *triage; changed = true }`.
   Verify: `go build ./...` passes.

6. **[cmd_show.go:showJSON]** — Add `Triage string \`json:"triage,omitempty"\`` field to `showJSON`, populate it in the JSON branch (`Triage: t.Triage`), and add `if t.Triage != "" { fmt.Printf("triage: %s\n", t.Triage) }` to the text output block (after `snooze`).
   Verify: `go build ./...` passes.

7. **[specs/ticket_triage.feature]** — Write a new spec covering: create with `--triage`, update with `--triage`, show displays the field, ticket without triage has no triage field in frontmatter.
   Verify: file is present and well-formed.

8. **[testdata/ticket_triage/triage_basic.txtar]** — Write a txtar integration test: create a ticket with `--triage "unblock this ticket"`, verify it appears in `ko show` output; update a ticket with `--triage "break this apart"`, verify `ko show` reflects the new value; create without `--triage`, verify no `triage:` line in output.
   Verify: `go test ./... -run TestScript` passes.

9. **[ticket_test.go]** — Add `TestTriageRoundTrip` (format → parse → check `t.Triage`) and `TestParseTicketWithTriage` (parse raw frontmatter string with `triage:` field). Follow the `TestSnoozeRoundTrip` / `TestParseTicketWithSnooze` pattern.
   Verify: `go test ./...` passes.

## Open Questions
None. `triage` is a free-text string with no validation — the pattern is direct and complete from the `snooze`/`assignee` precedents.
