Perfect! I've created a comprehensive plan for removing the `ko query` command. The plan includes:

**Goal**: Remove the redundant `ko query` command since list commands now support `--json`.

**Key findings**:
- The command outputs JSONL (one JSON per line) 
- The shared `ticketToJSON()` helper needs to be moved to `cmd_list.go` (not deleted)
- The SSE endpoint is safe - it uses `ListTickets()` directly, not the command
- Need to remove: source file, main.go dispatcher, help text, serve whitelist, tests, and spec

**9 ordered tasks** covering:
1. Delete cmd_query.go
2. Move ticketToJSON() to cmd_list.go 
3. Remove main.go switch case
4. Update help text
5. Remove serve whitelist entry
6. Delete test function
7. Delete test data
8. Delete spec file  
9. Verify all tests pass

The plan is ready for your review!
