The plan has been written to `.ko/tickets/ko-ff51.artifacts/plan.md`.

## Summary

After investigating the codebase, I found that **the desired behavior already exists**. The `statusOrder` function in `ticket.go` correctly ranks `in_progress` tickets (rank 0) ahead of `open` tickets (rank 1), and this is used throughout the sorting logic.

The plan focuses on:
1. **Adding a test** to explicitly verify and protect this behavior
2. **Updating documentation** in README.md to mention the status-based ordering within priority tiers

The implementation is already correct — this ticket is about ensuring the behavior is tested and documented rather than implementing new functionality.
