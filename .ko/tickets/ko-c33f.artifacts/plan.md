## Goal
Exclude snoozed tickets from `ko ready` (and the loop's ready queue) when the snooze date is in the future.

## Context
- `IsReady(status string, allDepsResolved bool) bool` in `ticket.go:85` is the pure decision function for ready-queue membership. It does not currently know about snooze.
- `Ticket.Snooze` field already exists (`ticket.go:34`); the snooze date is stored as a date-only string (`2006-01-02` format) and validated on write.
- `IsReady` is called in three places:
  - `cmd_list.go:228` — local ready queue in `cmdReady`
  - `cmd_list.go:274` — cross-project deps check in `cmdReady`
  - `loop.go:51` — `ReadyQueue` function used by the build loop
- `IsSnoozed` does not yet exist. Per INVARIANTS.md: "Decision logic is pure" and "New logic goes into testable functions first."
- Tests live in `ticket_test.go` (unit) and `testdata/ticket_snooze/*.txtar` (integration, run by `TestTicketSnooze` in `ko_test.go:68`).
- Specs live in `specs/ticket_snooze.feature`. New behavior needs a spec scenario before or alongside the implementation.
- Snooze date semantics per the ticket: "valid as of midnight on that date" — a ticket with `snooze: 2026-05-01` becomes ready at 2026-05-01T00:00:00Z (midnight UTC), consistent with how the rest of the codebase uses UTC.

## Approach
Add a pure `IsSnoozed(snooze string, now time.Time) bool` function to `ticket.go` that returns true when `snooze` is non-empty and `now` is before midnight UTC on the parsed snooze date. Update the three `IsReady` call sites to also gate on `!IsSnoozed(t.Snooze, time.Now())`. Add the spec scenarios and a txtar integration test.

## Tasks
1. **[specs/ticket_snooze.feature]** — Add two new scenarios: "Ready excludes a ticket snoozed until a future date" and "Ready includes a ticket whose snooze date has passed". This fulfills the "spec before code" invariant.
   Verify: file is valid Gherkin; no code changes yet.

2. **[ticket.go]** — Add `IsSnoozed(snooze string, now time.Time) bool` after `IsReady`. Parse `snooze` with `time.Parse("2006-01-02", snooze)` using UTC (`time.ParseInLocation`). Return `false` for empty or unparseable snooze strings (invalid dates are rejected at write time; silent no-op here is safe). Return `now.Before(parsedDate)` where `parsedDate` is midnight UTC on the snooze day.
   Verify: `go build ./...` passes.

3. **[ticket_test.go]** — Add `TestIsSnoozed` with cases: empty snooze (not snoozed), past date (not snoozed), future date (snoozed), today at midnight exactly (not snoozed — boundary: valid as of midnight), invalid string (not snoozed). Use a fixed `now` value for determinism.
   Verify: `go test ./... -run TestIsSnoozed` passes.

4. **[cmd_list.go:cmdReady]** — At the two `IsReady(...)` call sites (lines 228 and 274), add `&& !IsSnoozed(t.Snooze, time.Now())` to the condition. No other changes in this function.
   Verify: `go build ./...` passes.

5. **[loop.go:ReadyQueue]** — At line 51, add `&& !IsSnoozed(t.Snooze, time.Now())` to the `IsReady(...)` condition.
   Verify: `go build ./...` passes.

6. **[testdata/ticket_snooze/ready_exclusion.txtar]** — New txtar file exercising: (a) `ko ready` excludes a ticket with `snooze: 2099-01-01` (far future); (b) `ko ready` includes a ticket with `snooze: 2020-01-01` (past date). Use static ticket files in the txtar archive, not `ko add`, to control snooze values precisely without relying on `$TODAY`.
   Verify: `go test ./... -run TestTicketSnooze` passes, including the new txtar.

## Open Questions
None. The semantics ("valid as of midnight", UTC, date-only format) match the existing snooze validation convention. Invalid snooze strings are rejected at write time so silent skip in `IsSnoozed` is safe.
