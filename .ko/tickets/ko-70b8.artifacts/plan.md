## Goal
Rename the `ko reopen` command to `ko open` for consistency with the status name.

## Context
The `reopen` command is a convenience wrapper that sets ticket status to "open". The status is called "open", not "reopen", so the command name should match.

Key locations found:
- `main.go:34` — case statement for "reopen" command
- `main.go:131` — help text showing "reopen <id>        Set status to open"
- `cmd_status.go:72-78` — `cmdReopen()` function implementation
- `cmd_status.go:74` — error message "ko reopen: ticket ID required"
- `cmd_serve.go:335` — whitelist entry for HTTP server
- `cmd_serve_test.go:36, 205, 218` — test whitelist entries
- `specs/ticket_status.feature:41-45` — gherkin scenario testing reopen command
- `specs/http_server.feature:108` — HTTP server feature file mentioning reopen
- `testdata/ticket_status/status_shortcuts.txtar:11-14` — txtar test using reopen
- `README.md:24` — documentation showing the command

## Approach
Rename all occurrences of "reopen" to "open" throughout the codebase. The existing "open" status is already used internally; we're just changing the command name to match. This is a straightforward find-and-replace operation across Go source, tests, specs, and documentation.

## Tasks
1. [main.go:34] — Change case "reopen" to case "open" in command dispatcher.
   Verify: `go build` succeeds.

2. [main.go:131] — Update help text from "reopen <id>" to "open <id>".
   Verify: `go run . help` shows correct text.

3. [cmd_status.go:72-78] — Rename function from `cmdReopen` to `cmdOpen` and update error message from "ko reopen:" to "ko open:".
   Verify: `go build` succeeds.

4. [cmd_serve.go:335] — Change whitelist key from "reopen" to "open".
   Verify: `go build` succeeds.

5. [cmd_serve_test.go:36, 205, 218] — Update all three test whitelist entries from "reopen" to "open".
   Verify: `go test -run TestServe` passes.

6. [specs/ticket_status.feature:41-45] — Update scenario title and command from "reopen" to "open".
   Verify: spec reads correctly.

7. [specs/http_server.feature:108] — Change "reopen" to "open" in whitelisted subcommands table.
   Verify: spec reads correctly.

8. [testdata/ticket_status/status_shortcuts.txtar:11-14] — Update test command from "ko reopen" to "ko open" and comment from "# Reopen command" to "# Open command".
   Verify: `go test -run TestScript/ticket_status/status_shortcuts` passes.

9. [README.md:24] — Update documentation from "reopen <id>" to "open <id>".
   Verify: documentation is consistent.

## Open Questions
None. This is a straightforward rename with clear scope and no ambiguity.
