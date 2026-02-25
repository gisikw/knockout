# Summary: Ensure in_progress tickets sort ahead of open ones

## What Was Done

This ticket requested ensuring that `in_progress` tickets sort ahead of `open` tickets within the same priority tier. Upon investigation, the sorting logic in `ticket.go` (`SortByPriorityThenModified` and `statusOrder`) already implemented this behavior correctly:

- `statusOrder` assigns: `in_progress=0`, `open=1`, `resolved=2`, `closed=3`, `default=4`
- `SortByPriorityThenModified` sorts by: priority first, then status (via statusOrder), then ModTime descending

Rather than implementing new functionality, this ticket ensured and documented the existing behavior:

### 1. Added Comprehensive Test Coverage (ticket_test.go)

Created `TestSortByPriorityThenModified` with five test cases:
- Verifies `in_progress` sorts before `open` within same priority
- Confirms priority takes precedence over status
- Tests ModTime as the final tiebreaker
- Validates full ordering across multiple statuses
- Ensures `blocked` status sorts after `open`

All tests pass successfully.

### 2. Updated Documentation (README.md:88-93)

Updated the sorting documentation to explicitly state that within the same priority tier, tickets are sorted:
1. First by status (`in_progress` before `open`)
2. Then by file modification time (most recently touched first)

Previously, the documentation only mentioned ModTime sorting within priority tiers, omitting the status ordering that was already implemented.

## Notable Decisions

- Chose to add tests rather than modify code, since the implementation was already correct
- Test cases cover edge cases including mixed priorities, statuses, and timestamps
- Documentation change is minimal but makes the behavior explicit for users

## Implementation Notes

- No changes to sorting logic were needed
- The code comment at ticket.go:572-574 already documented the intent
- This was a verification and documentation task, not a new feature implementation
- All existing tests continue to pass

## Invariant Compliance

Reviewed INVARIANTS.md:
- ✅ Test added to verify behavior (Specifications and Tests section)
- ✅ No multi-purpose functions created (Code Organization section)
- ✅ Pure decision logic in sorting functions (Code Organization section)
- ✅ Test mirrors source file structure (ticket_test.go for ticket.go)

No invariants were violated.
