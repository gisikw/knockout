## Goal
Add `--limit` and `--json` flags to `ko ls` and related list commands (ready, blocked, resolved, closed) to output JSON arrays suitable for HTTP API responses.

## Context
The codebase already has:
- A `ticketJSON` struct in `cmd_query.go` that defines the JSON schema (id, title, status, type, priority, deps, created, modified, assignee, parent, tags)
- Existing `--json` flag support in all list commands that outputs JSONL (JSON Lines, one object per line)
- Existing `--limit` flag in `cmdClosed` and `cmdResolved` that truncates output after N items
- A `ko query` command that outputs all tickets as JSONL using the same schema
- The sorting happens via `SortByPriorityThenModified()` before filtering/output

The ticket requests:
1. `--json` to output as a JSON array (not JSONL) - this is a format change
2. `--limit N` to return only first N items after sorting - partially exists, needs to be added to ls, ready, blocked
3. Additional fields in JSON: `hasUnresolvedDep` (computed boolean) and `description` (ticket body) and `plan-questions`

Current `ticketJSON` struct is missing:
- `description` field (maps to `t.Body`)
- `hasUnresolvedDep` field (computed boolean)
- `plan-questions` field (maps to `t.PlanQuestions`)

The spec file `specs/ticket_listing.feature` documents that --json should output "valid JSONL", but the ticket explicitly requests JSON array format instead, which makes sense for HTTP API responses (most API clients expect JSON arrays, not JSONL).

## Approach
1. Extend `ticketJSON` struct to include the missing fields: `Description`, `HasUnresolvedDep`, and `PlanQuestions`.
2. Update all commands that use `ticketJSON` to populate these new fields when encoding.
3. Change `--json` output from JSONL to JSON array format across all list commands.
4. Add `--limit` flag to `cmdLs`, `cmdReady`, and `cmdBlocked` (already exists in `cmdClosed` and `cmdResolved`).
5. Update the spec and add corresponding tests.

## Tasks
1. [cmd_query.go:ticketJSON] — Add three new fields: `Description string`, `HasUnresolvedDep bool`, and `PlanQuestions []PlanQuestion` to the struct. Use appropriate json tags.
   Verify: `go test ./...` passes.

2. [cmd_query.go:cmdQuery] — Update the JSON encoding to populate `Description` from `t.Body`, compute `HasUnresolvedDep` using `!AllDepsResolved(ticketsDir, t.Deps)`, and copy `PlanQuestions` from `t.PlanQuestions`.
   Verify: `go test ./...` passes.

3. [cmd_list.go:cmdLs] — Change --json output from JSONL (line-by-line encoding) to JSON array. Collect all matching tickets in a slice, then `json.MarshalIndent` and write once. Also add `--limit` flag with same behavior as in `cmdClosed`.
   Verify: Manual test with `ko ls --json` outputs valid JSON array. Manual test with `ko ls --limit 2` returns 2 items.

4. [cmd_list.go:cmdReady] — Same changes as cmdLs: change to JSON array output and add `--limit` flag.
   Verify: Manual test with `ko ready --json` outputs valid JSON array.

5. [cmd_list.go:cmdBlocked] — Same changes: JSON array output and add `--limit` flag.
   Verify: Manual test with `ko blocked --json` outputs valid JSON array.

6. [cmd_list.go:cmdClosed] — Change JSONL to JSON array output. The --limit flag already exists.
   Verify: Manual test with `ko closed --json` outputs valid JSON array.

7. [cmd_list.go:cmdResolved] — Change JSONL to JSON array output. The --limit flag already exists.
   Verify: Manual test with `ko resolved --json` outputs valid JSON array.

8. [cmd_list.go] — Update all commands to populate the new fields (Description, HasUnresolvedDep, PlanQuestions) when creating ticketJSON instances.
   Verify: `go test ./...` passes.

9. [specs/ticket_listing.feature] — Update the spec scenarios that reference "valid JSONL" to say "valid JSON array" instead. Update test expectations to match array format.
   Verify: Spec accurately describes new behavior.

10. [testdata/ticket_listing/json_output.txtar] — Update the test to expect JSON array format with brackets and commas, not JSONL. Add assertions for the new fields (description, hasUnresolvedDep, plan-questions). Add test cases for --limit flag on ls, ready, and blocked commands.
    Verify: `go test ./...` passes with the new expectations.

## Open Questions
None — the ticket is explicit about requirements: JSON array format (not JSONL), add the three missing fields, and add --limit to all list commands. The implementation is straightforward refactoring of existing patterns.
