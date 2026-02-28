The plan is written to `.ko/tickets/ko-7a57.artifacts/plan.md`. Here's a summary:

**Goal**: Add `snooze: <ISO date>` to ticket frontmatter, with `--snooze` flags on `ko add` and `ko update`.

**10 tasks across 6 files:**
1. `ticket.go:Ticket` — add `Snooze string` field (omitempty)
2. `ticket.go:FormatTicket` — emit `snooze:` when non-empty
3. `ticket.go:ParseTicket` — parse the `snooze` key
4. `cmd_create.go` — `--snooze` flag with `2006-01-02` date validation
5. `cmd_update.go` — same flag on update
6–8. Tests in `ticket_test.go`, `cmd_create_test.go`, `cmd_update_test.go`
9. `specs/ticket_snooze.feature` — required by INVARIANTS.md ("every behavior has a spec")
10. Run `go test ./...`

**Two open questions flagged:**
- **Clearing snooze**: No mechanism to unset it after setting (consistent with how `assignee`/`external-ref` work today). Assumed out of scope.
- **Behavioral filtering**: Snoozed tickets will still appear in `ko list`/ready queue — only the data model is in scope here.
