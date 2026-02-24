I've written the implementation plan to `/home/dev/Projects/knockout/.ko/tickets/ko-84c5.artifacts/plan.md`.

The plan identifies:
- **Goal**: Add `--json` flag to output-producing commands (`ko show`, `ko ls`, `ko ready`, `ko blocked`, `ko closed`)
- **Context**: `ko query` already implements JSONL output using a `ticketJSON` struct as the reference pattern
- **Approach**: Extend the same JSONL pattern to other commands - single JSON object for `ko show`, one per line for list commands
- **Tasks**: 9 ordered tasks covering code changes, specs, and tests
- **Open Questions**: 3 questions about implementation details with stated assumptions for each:
  1. Include derived fields (blockers, blocking, children) in `ko show --json` 
  2. Include body content in `ko show --json` but not list commands
  3. Leave `ko query` unchanged as the reference implementation

The plan follows the project's invariants requiring specs before code and tests for every behavior change.
