# Implementation Summary: ko-57d3

## What Was Done

Successfully added `--limit` and `--json` flags to all list commands (ls, ready, blocked, resolved, closed) to make CLI output suitable for HTTP API responses.

### Changes Made

1. **Extended JSON schema** (`cmd_query.go:ticketJSON`):
   - Added `Description string` field (maps to `t.Body`)
   - Added `HasUnresolvedDep bool` field (computed via `AllDepsResolved()`)
   - Added `PlanQuestions []PlanQuestion` field (maps to `t.PlanQuestions`)

2. **Changed output format** from JSONL to JSON array:
   - All list commands now collect results in a slice
   - Use `json.MarshalIndent` with 2-space indentation for readability
   - Output wrapped in brackets `[...]` instead of newline-separated objects

3. **Added --limit flag** to ls, ready, and blocked commands:
   - Already existed in closed/resolved, now consistent across all commands
   - Applied after sorting by priority then modified time
   - Works in both JSON and human-readable output modes

4. **Updated specs and tests**:
   - `specs/ticket_listing.feature`: Changed "valid JSONL" to "valid JSON array"
   - `testdata/ticket_listing/json_output.txtar`: Added assertions for new fields and --limit behavior

### Implementation Decisions

- Used existing `AllDepsResolved()` helper for computing `hasUnresolvedDep` to maintain consistency with existing dependency resolution logic
- Applied --limit after sorting to ensure users get the highest-priority items first
- Preserved all existing behavior for non-JSON output mode
- Used `json.MarshalIndent` instead of line-by-line encoding for cleaner, more standard JSON array output

### Verification

- All tests pass (`go test ./...`)
- Manual verification confirms:
  - JSON array format with proper brackets and indentation
  - New fields populated correctly (description, hasUnresolvedDep, plan-questions)
  - --limit flag works on all list commands
  - hasUnresolvedDep correctly computed (true when ticket has unresolved deps)

## For Future Reference

The JSON output from `ko ls --json`, `ko ready --json`, etc. is now directly usable as HTTP API responses without any transformation layer. The schema matches `ko query` output but returns a JSON array instead of JSONL, making it compatible with standard REST API client expectations.

The `hasUnresolvedDep` field is always computed fresh on each query, ensuring it reflects the current state of dependencies even if tickets are resolved/closed between invocations.
