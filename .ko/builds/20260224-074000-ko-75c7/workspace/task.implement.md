Perfect! All tasks are complete. Let me provide a summary of the changes.

## Summary

I successfully completed all tasks in the plan to deprecate the link/unlink commands. Here's what was changed, organized by task:

### Code Changes

1. **main.go** - Removed `link` and `unlink` case branches from the command switch (lines 52-55)
2. **cmd_link.go** - Deleted entire file containing `cmdLink`, `cmdUnlink`, and helper functions
3. **ticket.go** - Removed all link-related code:
   - Removed `Links []string` field from `Ticket` struct
   - Removed links serialization line from `FormatTicket` 
   - Removed `Links: []string{}` initialization from `ParseTicket`
   - Removed `case "links"` from frontmatter parsing in `ParseTicket`
   - Removed `Links: []string{}` from `NewTicket` function
   - Removed `Links: []string{}` from `NewChildTicket` function
4. **cmd_show.go** - Removed `links: [...]` display line from `cmdShow`
5. **cmd_query.go** - Removed links from JSON output:
   - Removed `Links []string` field from `ticketJSON` struct
   - Removed `Links: t.Links` assignment in `cmdQuery`

### Test Changes

6. **ko_test.go** - Removed `TestTicketLinks` function
7. **testdata/ticket_links/** - Deleted entire directory
8. **specs/ticket_links.feature** - Deleted spec file
9. **All testdata/*.txtar files** - Removed all `links: []` lines from ticket frontmatter in test fixtures

### Documentation Changes

10. **main.go** (cmdHelp) - Removed `link <id1> <id2>` and `unlink <id1> <id2>` from help text
11. **README.md** - Removed `link <id1> <id2>` and `unlink <id1> <id2>` from usage section

### Verification

All changes have been verified:
- Code builds successfully with `go build .`
- Full test suite passes with `go test ./...`
- No references to links remain in help output or README

### Note on Deviations

The plan didn't explicitly call out the `NewTicket` and `NewChildTicket` functions in ticket.go, but they had `Links: []string{}` initializations that caused build failures. These were removed as necessary to complete the implementation. Additionally, test fixture files (.txtar) contained `links: []` in ticket frontmatter that needed to be cleaned up for tests to pass.

All changes are uncommitted as instructed.
