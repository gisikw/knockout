## Goal
Add a `snooze` field (ISO 8601 date string) to the Ticket data model, serialize/parse it, and expose it via `--snooze` flags on `ko add` and `ko update`.

## Context
- **`ticket.go`** holds the `Ticket` struct, `FormatTicket`, and `ParseTicket`. Optional string fields like `Assignee`, `ExternalRef` use `omitempty` and are emitted in `FormatTicket` only when non-empty. `ParseTicket` has a manual switch on frontmatter keys. `ticket.go` is already at 672 lines (over the 500-line invariant limit — additions here are unavoidable but should be minimal).
- **`cmd_create.go:cmdCreate`** parses flags and builds a new ticket. The `reorderArgs` call lists all flag names that take values. New flags must be added to that map. Validates inputs (e.g. priority must be ≥ 0), sets ticket fields, then calls `SaveTicket`.
- **`cmd_update.go:cmdUpdate`** follows the same pattern: `reorderArgs`, flag set, load ticket, apply non-zero flags, save. The `changed` bool gates the save.
- **INVARIANTS.md** requires: every new behavior gets a spec in `specs/*.feature` before or alongside implementation, and a corresponding Go test. Statuses are a closed set — snooze does not add a new status. No external dependencies.
- **Testing pattern**: unit tests use `t.TempDir()`, write a ticket file with `SaveTicket`, `os.Chdir` into the tmp dir, call the `cmd*` function directly, then reload and assert fields. `ticket_test.go` uses table-driven parse/roundtrip tests.

## Approach
Add `Snooze string` to the `Ticket` struct with `yaml:"snooze,omitempty"`, emit and parse it alongside the other optional fields. Expose `--snooze` on `ko add` and `ko update`, validating that the provided value is a parseable `2006-01-02` date. Write a spec and tests covering set-via-create, set-via-update, and roundtrip serialization. Do not implement behavioral side-effects (filtering snoozed tickets from `ready`/`list`) — that's out of scope for this ticket.

## Tasks

1. **[ticket.go:Ticket]** Add `Snooze string \`yaml:"snooze,omitempty"\`` field to the struct, after `ExternalRef` (keeping optional fields grouped).
   Verify: `go build ./...` passes.

2. **[ticket.go:FormatTicket]** After the `ExternalRef` block, add a symmetric `if t.Snooze != "" { b.WriteString(...) }` that emits `snooze: <value>\n`.
   Verify: `go build ./...` passes.

3. **[ticket.go:ParseTicket]** Add `case "snooze": t.Snooze = val` to the frontmatter key switch (alongside `external-ref`, `tags`, etc.).
   Verify: `go build ./...` passes.

4. **[cmd_create.go:cmdCreate]** Add `"snooze": true` to the `reorderArgs` map; add a `snooze := fs.String("snooze", "", "snooze date (ISO 8601, e.g. 2026-05-01)")` flag; after flag parsing, if `*snooze != ""`, validate it with `time.Parse("2006-01-02", *snooze)` (error out on bad format), then set `t.Snooze = *snooze`.
   Verify: `go build ./...` passes; `ko add "test" --snooze 2026-05-01` creates ticket with `snooze: 2026-05-01`; `ko add "test" --snooze bad-date` exits non-zero.

5. **[cmd_update.go:cmdUpdate]** Add `"snooze": true` to the `reorderArgs` map; add a `snooze := fs.String("snooze", "", "snooze date (ISO 8601, e.g. 2026-05-01)")` flag; after the other field updates, if `*snooze != ""`, validate with `time.Parse("2006-01-02", *snooze)`, set `t.Snooze = *snooze`, and set `changed = true`.
   Verify: `go build ./...` passes; `ko update <id> --snooze 2026-05-01` updates the field; `ko update <id> --snooze bad` exits non-zero.

6. **[ticket_test.go]** Add a `TestSnoozeRoundTrip` test: construct a `Ticket` with `Snooze: "2026-05-01"`, call `FormatTicket`, assert the output contains `snooze: 2026-05-01`, then `ParseTicket` the result and assert `Snooze == "2026-05-01"`. Also add a parse case to the existing parse table for a ticket with a snooze field.
   Verify: `go test ./... -run TestSnooze` passes.

7. **[cmd_create_test.go]** Add a test case for `ko add "Ticket" --snooze 2026-05-01` that asserts the saved ticket has `Snooze == "2026-05-01"`, and a case for an invalid date that expects a non-zero exit code.
   Verify: `go test ./... -run TestCmdCreate` passes.

8. **[cmd_update_test.go]** Add a test case for `ko update <id> --snooze 2026-05-01` that asserts the saved ticket has `Snooze == "2026-05-01"`, and a case for an invalid date that expects a non-zero exit code.
   Verify: `go test ./... -run TestCmdUpdate` passes.

9. **[specs/ticket_snooze.feature]** Write a Gherkin spec covering: (a) create a ticket with `--snooze`, verify frontmatter contains `snooze: <date>`; (b) update a ticket with `--snooze`, verify field updated; (c) invalid date format rejected. This satisfies the INVARIANTS.md "every behavior has a spec" requirement.
   Verify: Spec exists and is coherent with the implementation.

10. **[Final]** Run `go test ./...` to confirm all tests pass.

## Open Questions

1. **Clearing snooze via `ko update`**: The current flag pattern treats empty string as "no change." There's no way to clear `snooze` once set (e.g., `--snooze ""` won't work because the empty-check guards the update). This is consistent with how `assignee` and `external-ref` work today (also not clearable). Assumption: clearing is out of scope; a follow-up ticket can add `--clear-snooze` if needed.

2. **Behavioral effect on `ko list` / `ko ready`**: Snooze implies "hide until this date." The ticket only asks for the data model and CLI flags — it says nothing about filtering. Assumption: no filtering is implemented here. A snoozed ticket remains visible in `ko list` and eligible for `ready`. Future ticket to add filtering.
