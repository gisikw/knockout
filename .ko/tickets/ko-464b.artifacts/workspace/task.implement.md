## Summary

**Task 1 — `cmd_list.go:230`**
Added `&& t.Triage == ""` to the local ready queue filter. Build passes.

**Task 2 — `cmd_list.go:276`**
Added `&& t.Triage == ""` to the cross-project ready queue filter. Build passes.

**Task 3 — `ticket_test.go`**
Added `TestReadyTriageExclusion` documenting that the triage predicate (`t.Triage == ""`) is applied at the call site, not inside `IsReady`. All tests pass.

**Deviations from plan (INVARIANTS.md compliance)**

The plan only called for a `ticket_test.go` unit test. INVARIANTS.md requires every new behavior to have a spec and a corresponding testscript test. Following that requirement, I also:

- Added two `ko ready` triage exclusion scenarios to `specs/ticket_triage.feature` (mirroring the snooze spec pattern)
- Created `testdata/ticket_triage/ready_exclusion.txtar` (mirroring `testdata/ticket_snooze/ready_exclusion.txtar` exactly)
