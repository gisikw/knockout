Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-dd51.artifacts/plan.md`.

The plan is ready for your review. It covers:

1. **Adding the status**: Insert "resolved" into the Statuses slice between "blocked" and the end
2. **Excluding from ready queue**: Update IsReady() to treat "resolved" like "closed" 
3. **Sort ordering**: Position "resolved" between active work and fully closed tickets
4. **New command**: Add `ko resolved` to list tickets needing review
5. **Tests**: Unit test to verify resolved tickets don't appear in ready queue

The implementation follows existing patterns in the codebase and is straightforward — no architectural decisions or open questions remain. Once approved, this can proceed directly to implementation.
