## Goal
Ensure `in_progress` tickets sort ahead of `open` tickets within the same priority tier.

## Context
Found the sorting logic in `ticket.go`:
- `SortByPriorityThenModified` (lines 558-570) sorts tickets by priority first, then by status via `statusOrder`, then by ModTime descending
- `statusOrder` (lines 575-588) assigns sort ranks to statuses: in_progress=0, open=1, resolved=2, closed=3, default=4
- This function is already correct — it ranks in_progress (0) before open (1)
- The sorting is used in:
  - `cmd_list.go:cmdLs` (line 74)
  - `cmd_list.go:cmdReady` (line 173) — for local ready queue
  - `loop.go:ReadyQueue` (line 53) — for build loop queue

Checked README.md (lines 88-93): documentation states tickets are sorted by ModTime within priority tier, but doesn't mention status ordering. The documentation accurately reflects the _existing_ behavior but doesn't document the status ordering that already exists in the code.

The code comment at ticket.go:572-574 already documents the intended behavior: "in_progress < open < blocked/etc < resolved < closed" and "Tickets with status 'open' but unresolved deps are still 'open' here — they sort with other open tickets, not with explicitly blocked ones."

## Approach
The sorting logic already implements the desired behavior (in_progress before open). This appears to be a documentation request or a verification that the behavior is correct. However, looking at the ticket title "Ensure in_progress tickets sort ahead of open ones", this suggests either:
1. The behavior exists but needs testing
2. The documentation should be updated to reflect this behavior

Given "medium discretion", I'll verify the implementation is correct (it is), add a test to ensure the behavior is explicit and protected, and update documentation if needed.

## Tasks
1. [ticket_test.go] — Add test `TestSortByPriorityThenModified` to verify in_progress tickets sort before open tickets within the same priority tier.
   Verify: `go test -run TestSortByPriorityThenModified ./...` passes.

2. [README.md:88-93] — Update the status sorting documentation to explicitly mention that in_progress tickets appear before open tickets within the same priority tier.
   Verify: documentation accurately reflects the code behavior.

## Open Questions
None — the implementation already provides the desired behavior. This ticket appears to be about ensuring (via tests) and documenting the existing behavior rather than implementing new functionality.
