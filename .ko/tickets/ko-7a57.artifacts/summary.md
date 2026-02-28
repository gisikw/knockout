## Summary

### What was done

The implementation correctly added the `snooze` field (ISO 8601 date string) to the ticket data model and exposed it via `--snooze` flags on `ko add` and `ko update`. All 10 planned tasks were completed:

- `Ticket.Snooze string` field added to the struct with `yaml:"snooze,omitempty"`
- `FormatTicket` emits `snooze: <date>` when non-empty
- `ParseTicket` handles `case "snooze"` in the frontmatter key switch
- `ko add --snooze <date>` validates ISO 8601 format and sets the field
- `ko update --snooze <date>` validates and updates the field
- Unit tests in `ticket_test.go`, `cmd_create_test.go`, `cmd_update_test.go`
- `specs/ticket_snooze.feature` Gherkin spec covering all four behaviors

### Fixes applied during review

Three gaps were found and corrected:

1. **Missing testscript tests** — INVARIANTS.md requires `testdata/*.txtar` files as the verification layer ("A spec without a corresponding test is an unverified claim"). The implementation only added `*_test.go` unit tests. Added `testdata/ticket_snooze/snooze_basic.txtar` and `TestTicketSnooze` in `ko_test.go` to satisfy this invariant.

2. **`cmd_show.go` missing snooze** — The `ko show` text output and `showJSON` struct both omitted the `snooze` field, making the value invisible to any consumer of `ko show`. Added `snooze:` to the text frontmatter display and `Snooze` to `showJSON`.

3. **`cmd_list.go` ticketJSON missing snooze** — Same omission: `ko ls --json` silently dropped the snooze field. Added `Snooze string \`json:"snooze,omitempty"\`` to `ticketJSON` and populated it in `ticketToJSON`. This was also necessary to write meaningful testscript assertions.

### Notable decisions

- **No behavioral side-effects**: Snoozed tickets remain visible in `ko list` and eligible for `ko ready`. The ticket only asked for the data model and CLI flags; filtering is deferred to a future ticket. (Confirmed by plan's open question #2.)
- **No clearing mechanism**: `--snooze ""` does not clear the field, consistent with how `assignee` and `external-ref` work today. (Confirmed by plan's open question #1.)
- **`ticket.go` line count**: Now at 678 lines, over the 500-line invariant limit. The plan acknowledged this as unavoidable; a split is tracked separately.

### Out-of-scope changes in diff

The diff also includes unrelated pipeline state changes (`.ko/pipeline.yml` model bump, ticket status transitions for `ko-88dd` and `ko-dc03`). These are pipeline bookkeeping, not part of this feature.
