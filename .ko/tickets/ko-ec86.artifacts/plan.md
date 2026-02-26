## Goal
Add an `--all` flag to `ko ls` that includes closed tickets, matching the behavior of SSE endpoints which return all tickets by default.

## Context
Currently, `ko ls` filters out closed tickets by default (cmd_list.go:158-160). Users must explicitly use `--status=closed` to see closed tickets. However, the SSE endpoints (`/subscribe/` and `/status/`) in `ko serve` return all tickets including closed ones (cmd_serve_sse.go:196-215).

The ticket requests parity with the SSE behavior - adding an `--all` flag that disables the default closed-ticket filter, allowing `ko ls --all` to show all tickets regardless of status.

Key files:
- cmd_list.go:122-193 - The `cmdLs` function containing the filtering logic
- cmd_list.go:158-160 - The specific lines that filter closed tickets
- cmd_serve_sse.go:196-215 - The SSE behavior that shows all tickets (for reference)
- specs/ticket_listing.feature - Behavioral specifications for listing

Project conventions (from INVARIANTS.md):
- Every behavior needs a spec (specs/*.feature)
- Every spec needs a test (testdata/*.txtar)
- Functions that make decisions should be pure (take data, return decisions)
- 500 line max per file (cmd_list.go is currently 286 lines)

## Approach
Add an `--all` boolean flag to `cmdLs` that, when set, disables the default closed-ticket filter. The implementation is straightforward: add the flag to the FlagSet, then check it before skipping closed tickets. This maintains backward compatibility (default behavior unchanged) while providing the requested parity with SSE behavior.

## Tasks
1. [cmd_list.go:cmdLs] — Add `--all` flag to the FlagSet after line 138. Add the filtering logic check: when `*allTickets` is false, apply the existing closed-ticket filter; when true, skip the filter.
   Verify: `ko ls` shows only non-closed tickets (existing behavior), `ko ls --all` shows all tickets including closed ones.

2. [specs/ticket_listing.feature] — Add a scenario under the "List with status filter" section that validates `ko ls --all` includes closed tickets.
   Verify: New spec documents the expected behavior.

3. [testdata/ticket_listing.txtar] — Add a testscript test case that creates open and closed tickets, runs `ko ls --all`, and asserts both appear in output.
   Verify: `go test ./... -run TestScript/ticket_listing` passes for the new test case.

## Open Questions
None. The implementation is straightforward and the requirement is clear: add `--all` flag to match SSE behavior of including closed tickets. The default behavior (excluding closed) is preserved for backward compatibility.
