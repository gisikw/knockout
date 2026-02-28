# Summary

## What was done

`ko ready` now excludes tickets that have a non-empty `triage` field. The change follows the same pattern as the snooze exclusion.

### Changes

- **`cmd_list.go`** (lines 230 and 276): Added `&& t.Triage == ""` to both ready-queue filter conditions — the local queue and the cross-project queue. Both sites now read:
  `IsReady(...) && !IsSnoozed(...) && t.Triage == ""`

- **`ticket_test.go`**: Added `TestReadyTriageExclusion` documenting the predicate. The test is minimal by design — the check is `t.Triage == ""` with no helper function needed. The test explicitly notes that triage exclusion is NOT handled inside `IsReady`.

- **`specs/ticket_triage.feature`**: Two new scenarios covering the exclusion (ticket with triage omitted, ticket with triage included).

- **`testdata/ticket_triage/ready_exclusion.txtar`**: Integration test mirroring the snooze pattern (`ready_exclusion.txtar` in `ticket_snooze`). Fixtures include a seed ticket, a triaged ticket (`ko-triaged`), and a plain ticket (`ko-plain`). The `TestTicketTriage` testscript runner in `ko_test.go` picks it up automatically.

## Notable decisions

- No helper function (`IsTriaged` or similar) was introduced. The check is a one-liner (`t.Triage == ""`), and the codebase precedent (snooze uses `IsSnoozed` because date logic is non-trivial) doesn't apply here. A named function would be over-engineering.
- The `ticket_test.go` unit test tests the predicate directly rather than through a helper. This is acknowledged as somewhat tautological but documents the intended behavior and the design decision to keep triage exclusion out of `IsReady`.

## All tests pass

`go test ./...` passes cleanly.
