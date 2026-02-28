## Goal
Exclude tickets with a non-empty `triage` field from the `ko ready` output.

## Context
`IsReady` (`ticket.go:86`) is a pure function taking `status` and `allDepsResolved`; it has no knowledge of snooze or triage. The snooze exclusion follows a different pattern: `IsSnoozed` is a separate pure function, and the actual filtering happens at two call sites in `cmd_list.go` (lines 230 and 276) with `&& !IsSnoozed(t.Snooze, time.Now())`.

The `Triage` field (`ticket.go:35`) is already parsed, formatted, and round-trip tested. Excluding a ticket means its `Triage` string is non-empty. There is no date logic — it's a simple presence check.

Tests for `IsReady` live in `ticket_test.go:9`. There are no dedicated integration tests for `ko ready` filtering.

## Approach
Add `&& t.Triage == ""` at both `ko ready` filter sites in `cmd_list.go`, mirroring the snooze pattern. Add test cases to `TestIsReady` — actually triage doesn't flow through `IsReady`, so instead add a new `TestReadyExcludesTriage` table test directly against the cmd_list filtering logic, or add cases to the existing `ticket_test.go` that document the exclusion. The cleanest fit is to add test cases to `ticket_test.go` testing a helper, but since the check is just `t.Triage == ""` with no helper needed, add the test inline to document the behavior.

## Tasks
1. [`cmd_list.go:230`] — Add `&& t.Triage == ""` to the local ready queue filter condition.
   Verify: `go build ./...` passes.

2. [`cmd_list.go:276`] — Add `&& t.Triage == ""` to the cross-project ready queue filter condition.
   Verify: `go build ./...` passes.

3. [`ticket_test.go`] — Add test cases to `TestIsReady` documenting that triage exclusion is NOT handled by `IsReady` (so no change needed there), and add a new `TestReadyTriageExclusion` test that exercises the filtering logic used in `cmd_list.go` by constructing tickets with and without a triage value and asserting the expected `t.Triage == ""` predicate.
   Verify: `go test ./...` passes with new tests green.

## Open Questions
None. The ticket title and the snooze precedent make the behavior unambiguous: any non-empty `triage` string means the ticket is excluded from `ko ready`.
