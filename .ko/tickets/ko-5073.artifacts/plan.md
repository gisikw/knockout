## Goal
Make `ko triage` and `ko agent triage` resolve a ticket's project from the ID prefix (e.g., `fn-a001`) via the registry, rather than failing when the caller isn't inside a project directory.

## Context

**How `ko close` works:** `cmdClose` delegates immediately to `cmdUpdate` with no project-resolution guard of its own. `cmdUpdate` calls `resolveProjectTicketsDir([]string{ticketID})` to get a local `ticketsDir` (may be `""`), then `ResolveTicket(ticketsDir, ticketID)` which tries local first and falls back to registry-based prefix lookup if the ticket isn't found locally.

**How `ko triage` currently breaks:** `cmdTriage` calls `resolveProjectTicketsDir(args)` and then immediately fails with "no .ko/tickets directory found" if `ticketsDir == ""` — even when the caller just wants to set triage on a cross-project ticket (e.g., `ko triage fn-a001 "do something"`). This guard executes before the flag parsing that reveals whether a ticket ID was supplied.

**Relevant code:**
- `cmd_list.go:292–360` — `cmdTriage`: the early `ticketsDir == ""` check is the bug.
- `cmd_agent_triage.go:12–78` — `cmdAgentTriage`: same early check, same problem.
- `ticket.go:506–543` — `ResolveTicket`: already handles cross-project prefix lookup when local lookup fails; works correctly with `ticketsDir == ""` (falls through to registry).
- `cmd_list_test.go:266–331` — existing `cmdTriage` tests (all run inside a temp project dir).

## Approach

Move the `if ticketsDir == ""` guard in `cmdTriage` to after the flag-parsing block, scoped only to the listing branch (zero positional args). The update branch (`NArg >= 2`) already delegates to `cmdUpdate`, which handles cross-project resolution. Apply the same fix to `cmdAgentTriage`: remove its early `ticketsDir == ""` guard and let `ResolveTicket` handle cross-project lookup.

## Tasks

1. **[cmd_list.go:cmdTriage]** — Move the `if ticketsDir == ""` check from its current position (before `fs.Parse`) to the listing branch (after `fs.NArg() == 1` error). Only the listing path (zero positional args) needs a local project directory.
   Verify: `go test ./... -run TestCmdTriage` passes.

2. **[cmd_agent_triage.go:cmdAgentTriage]** — Remove the `if ticketsDir == ""` early-exit block. `ResolveTicket` is already called on line 40 and handles the cross-project case; no guard needed.
   Verify: `go test ./... -run TestCmdAgentTriage` passes (if that test exists); broader `go test ./...` passes.

3. **[cmd_list_test.go]** — Add a test case to `TestCmdTriage` (or a new `TestCmdTriageCrossProject` sub-test) that:
   - Sets up two project directories in a registry: one with a different prefix (e.g., `fn`) containing a ticket `fn-test`, another as the "current" project.
   - Changes into a temp dir that is NOT either project (or into the current project dir).
   - Calls `cmdTriage([]string{"fn-test", "do something"})` and asserts it returns 0 and the triage field is set on `fn-test`.
   Verify: new test passes; existing tests unbroken.

## Open Questions

None. The fix is mechanical: the listing guard is in the wrong place.
