## Goal
Add `ko serve` HTTP daemon with single `/ko` endpoint for executing whitelisted ko subcommands.

## Context
The knockout codebase follows a strict command pattern:
- All commands are registered in main.go's run() function via a switch statement
- Commands are implemented in cmd_*.go files (e.g., cmd_status.go, cmd_query.go, cmd_list.go)
- The codebase uses exec.Command in several places (harness.go, adapter.go, build.go, cmd_agent.go)
- No HTTP server currently exists in the codebase (confirmed via grep for "net/http")
- Tests use testscript pattern in testdata/ directories (see ko_test.go)
- INVARIANTS.md requires zero external runtime dependencies and files under 500 lines
- The binary path is available via os.Args[0] for self-execution
- Error messages go to stderr with non-zero exit codes
- The project uses flag package for parsing flags with reorderArgs helper

## Approach
Create cmd_serve.go with an HTTP server listening on configurable port (default :9876, -p flag). Single POST /ko endpoint accepts JSON body with argv array, validates against a whitelist of allowed subcommands, executes via exec.Command using os.Args[0], and returns stdout (200) or stderr (400) based on exit code. Graceful shutdown via signal handling. No authentication, SSE, WebSocket, CORS, or TLS per ticket requirements.

## Tasks
1. [cmd_serve.go] — Create new file with cmdServe function that:
   - Parses -p flag for port (default "9876")
   - Defines whitelist slice: ["ls", "ready", "blocked", "resolved", "closed", "query", "show", "questions", "answer", "close", "reopen", "block", "start", "bump", "note", "status", "dep", "undep", "agent"]
   - Sets up HTTP server with single POST /ko handler
   - Handler parses JSON body {"argv": ["subcommand", ...]}
   - Validates first argv element is in whitelist (return 400 if not)
   - Executes exec.Command(os.Args[0], argv...) with combined output
   - Returns 200 + stdout on exit 0, 400 + JSON error with stderr on non-zero
   - Implements graceful shutdown on SIGTERM/SIGINT
   - Returns 0 on success, 1 on failure (follows CLI convention)
   Verify: `go build` succeeds, file is under 500 lines.

2. [main.go:run] — Add case "serve": return cmdServe(rest) to switch statement (alphabetically after "reopen", before "show").
   Verify: Pattern matches existing command registration.

3. [main.go:cmdHelp] — Add serve command to help text after "reopen" line with description "  serve [-p port]    Start HTTP daemon (default :9876)".
   Verify: Formatting matches existing help entries.

4. [cmd_serve_test.go] — Create basic Go unit test that:
   - Tests whitelist validation (valid subcommands accepted, invalid rejected)
   - Tests JSON parsing (valid and invalid payloads)
   - Uses httptest.NewServer for testing HTTP handler
   - Verifies 200 vs 400 status codes based on subcommand validity
   Verify: `go test -run TestServe` passes.

5. [testdata/serve/basic.txtar] — Create testscript integration test that:
   - Initializes a ko project
   - Starts `ko serve` in background with custom port
   - Uses curl or similar to POST to /ko with valid and invalid subcommands
   - Verifies response codes and output
   - Stops the server gracefully
   Verify: `go test -run TestServe` passes.

6. [ko_test.go] — Add `func TestServe(t *testing.T) { testscript.Run(t, testscript.Params{Dir: "testdata/serve"}) }` after TestTicketQuery.
   Verify: Test function follows existing pattern.

## Open Questions
1. Should the whitelist include "create" and "add" subcommands? The ticket doesn't list them, but they're valid ko commands. Excluding them for now since the ticket provides an explicit list and doesn't mention create/add. If ticket creation via HTTP is needed, that should be a separate decision.

2. Should the server log requests? The ticket says "stderr → ignored (or logged)" for command stderr. Interpreting this as: command stderr goes into the 400 error response body when exit code is non-zero, but we don't log to a separate file. Server-level logging (requests, startup/shutdown) will go to stdout/stderr per Go http.Server defaults.

3. Should Content-Type validation enforce application/json on requests? Not mentioned in ticket, but standard practice. Will accept any Content-Type for now and rely on JSON unmarshaling to fail on invalid payloads (returns 400).

4. The ticket says "use the same binary path (os.Args[0])" but doesn't specify working directory for exec.Command. Will use current working directory (inherited from server process) since ko commands resolve project locations relative to cwd.
