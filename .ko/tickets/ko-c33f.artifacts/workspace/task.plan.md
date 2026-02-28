The plan is written to `.ko/tickets/ko-c33f.artifacts/plan.md`. Here's a summary:

**Goal:** Exclude tickets with a future snooze date from `ko ready` and the loop's ready queue.

**What I found:**
- `IsReady()` in `ticket.go` is the pure decision function — doesn't know about snooze
- `Ticket.Snooze` field already exists, populated during parse
- `IsReady` is called in 3 places: twice in `cmdReady` (cmd_list.go:228, 274) and once in `ReadyQueue` (loop.go:51)
- No `IsSnoozed` function exists yet

**Approach:** Add a pure `IsSnoozed(snooze string, now time.Time) bool` to `ticket.go`, gate all three `IsReady` call sites on `!IsSnoozed(t.Snooze, time.Now())`, and add the spec + txtar integration test.

**Tasks (6 steps):**
1. Add spec scenarios to `specs/ticket_snooze.feature`
2. Add `IsSnoozed` to `ticket.go`
3. Add `TestIsSnoozed` to `ticket_test.go`
4. Update `cmd_list.go` (both `IsReady` call sites)
5. Update `loop.go` (`ReadyQueue`)
6. Add `testdata/ticket_snooze/ready_exclusion.txtar`

No open questions — semantics are unambiguous from the ticket and existing conventions.
