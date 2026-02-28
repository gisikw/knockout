Plan written to `.ko/tickets/ko-464b.artifacts/plan.md`.

**Summary:**

The fix is small and follows the existing snooze pattern exactly. Two call sites in `cmd_list.go` (lines 230 and 276) filter the ready queue with `IsReady(...) && !IsSnoozed(...)`. Both need `&& t.Triage == ""` appended.

`IsReady` itself doesn't need to change — it's a pure status/deps function, and triage exclusion is handled at the call site, just like snooze.

**Tasks:**
1. `cmd_list.go:230` — add `&& t.Triage == ""` to the local ready filter
2. `cmd_list.go:276` — add `&& t.Triage == ""` to the cross-project ready filter
3. `ticket_test.go` — add a `TestReadyTriageExclusion` test documenting the behavior

**Open questions:** None.
