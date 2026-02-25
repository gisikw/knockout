I've created a comprehensive plan for implementing the `--limit` and `--json` flags. Here's what I found and planned:

## Key Findings

1. **Existing Infrastructure**: The codebase already has `--json` support in all list commands, but it outputs JSONL (JSON Lines) format. The ticket explicitly requests JSON array format instead.

2. **Partial --limit Support**: The `--limit` flag already exists in `cmdClosed` and `cmdResolved`, but needs to be added to `cmdLs`, `cmdReady`, and `cmdBlocked`.

3. **Missing Fields**: The current `ticketJSON` struct is missing three fields the punchlist client needs:
   - `description` (the ticket body)
   - `hasUnresolvedDep` (computed boolean indicating if the ticket has unresolved dependencies)
   - `plan-questions` (for tickets that need human input before implementation)

4. **Testing Pattern**: The project uses gherkin specs in `specs/*.feature` with corresponding `testscript` tests in `testdata/*.txtar`. Both will need updates.

## Implementation Approach

The plan breaks down into 10 tasks that:
1. Extend the JSON schema with the missing fields
2. Convert all commands from JSONL to JSON array output
3. Add `--limit` flags where missing
4. Update specs and tests to match new behavior

All changes are incremental modifications to existing patterns — no architectural decisions required.
