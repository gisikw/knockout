## Summary

Three changes made, all tests passing:

**Task 1 — `cmd_list.go:cmdTriage`**  
Removed the early `if ticketsDir == ""` guard (lines 300–303) and moved it to just before the listing branch (`ListTickets` call). The update branch (`NArg >= 2`) already delegates to `cmdUpdate`, which handles cross-project resolution through `ResolveTicket`.

**Task 2 — `cmd_agent_triage.go:cmdAgentTriage`**  
Removed the `if ticketsDir == ""` early-exit block (lines 18–21). `ResolveTicket` is called on line 40 (now line 37) and handles the cross-project case by falling through to the registry when `ticketsDir == ""`.

**Task 3 — `cmd_list_test.go`**  
Added `TestCmdTriageCrossProject` — sets up a registry with an `fn` prefix pointing to a temp project dir containing `fn-test`, changes into an unrelated temp dir, calls `cmdTriage([]string{"fn-test", "do something"})`, and asserts it returns 0 with the triage field saved correctly.

No deviations from the plan.
