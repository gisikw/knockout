All 10 tasks complete. Here's a summary of what was changed:

---

### Summary

**`ticket.go`**
- Added `Snooze string \`yaml:"snooze,omitempty"\`` to `Ticket` struct (after `ExternalRef`)
- `FormatTicket`: added `if t.Snooze != "" { b.WriteString("snooze: ...") }` block after `external-ref`
- `ParseTicket`: added `case "snooze": t.Snooze = val` to the frontmatter switch

**`cmd_create.go`**
- Added `"snooze": true` to the `reorderArgs` map
- Added `snooze := fs.String("snooze", "", ...)` flag
- Added validation (`time.Parse("2006-01-02", ...)`) and assignment of `t.Snooze`

**`cmd_update.go`**
- Added `"time"` import
- Added `"snooze": true` to the `reorderArgs` map
- Added `snooze := fs.String("snooze", "", ...)` flag
- Added validation and assignment of `t.Snooze` with `changed = true`

**`ticket_test.go`**
- Added `"strings"` import
- Added `TestSnoozeRoundTrip`: constructs a ticket with `Snooze`, formats, checks output contains `snooze: 2026-05-01`, parses back, asserts field value
- Added `TestParseTicketWithSnooze`: parses a raw ticket string with `snooze:` frontmatter field

**`cmd_create_test.go`**
- Added `TestCreateWithSnooze`: valid date subtest (asserts `Snooze == "2026-05-01"`) and invalid date subtest (expects non-zero exit)

**`cmd_update_test.go`**
- Added `TestCmdUpdateSnooze`: valid date case (asserts field saved) and invalid date case (expects non-zero exit)

**`specs/ticket_snooze.feature`**
- New Gherkin spec with 5 scenarios covering: create with snooze, update with snooze, invalid date on create, invalid date on update, no snooze field when not set

**No deviations from the plan.**
