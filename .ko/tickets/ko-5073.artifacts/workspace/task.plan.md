Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-5073.artifacts/plan.md`.

**Summary:**

The bug is a misplaced guard in `cmdTriage` (and identically in `cmdAgentTriage`). Both functions call `resolveProjectTicketsDir` then immediately fail if `ticketsDir == ""` — before parsing flags or checking whether a ticket ID was supplied. This means `ko triage fn-a001 "instructions"` fails if you're not inside a project directory, even though the delegation to `cmdUpdate` (which does the actual save) already has full cross-project resolution via `ResolveTicket`.

**Three tasks:**
1. `cmd_list.go` — move the `ticketsDir == ""` check into the listing branch only (after `NArg` is known)
2. `cmd_agent_triage.go` — remove the same early-exit guard; `ResolveTicket` already handles the empty-dir case
3. `cmd_list_test.go` — add a cross-project test for `cmdTriage` that exercises the registry prefix lookup path
