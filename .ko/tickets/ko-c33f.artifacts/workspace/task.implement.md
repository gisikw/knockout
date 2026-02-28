All 6 tasks complete. Here's a summary of what was changed:

---

## Summary

**Task 1 — `specs/ticket_snooze.feature`**
Added two new Gherkin scenarios: "Ready excludes a ticket snoozed until a future date" and "Ready includes a ticket whose snooze date has passed".

**Task 2 — `ticket.go`**
Added `IsSnoozed(snooze string, now time.Time) bool` after `IsReady`. Uses `time.ParseInLocation("2006-01-02", snooze, time.UTC)` to parse the date as midnight UTC. Returns `false` for empty/unparseable input; returns `now.Before(parsedDate)` otherwise.

**Task 3 — `ticket_test.go`**
Added `TestIsSnoozed` with 5 cases using a fixed `now` (2026-03-15T12:00:00Z): empty snooze, past date, future date, today-at-midnight boundary (not snoozed), and invalid string.

**Task 4 — `cmd_list.go`**
At both `IsReady(...)` call sites (lines 228 and 274), added `&& !IsSnoozed(t.Snooze, time.Now())`.

**Task 5 — `loop.go`**
At the `ReadyQueue` call site (line 51), added `&& !IsSnoozed(t.Snooze, time.Now())`.

**Task 6 — `testdata/ticket_snooze/ready_exclusion.txtar`**
New txtar with static ticket files: `ko-future` (snooze: 2099-01-01) excluded from `ko ready`, `ko-past` (snooze: 2020-01-01) included.

**No deviations from the plan.** All verifications passed.
