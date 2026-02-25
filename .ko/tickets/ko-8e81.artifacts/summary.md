# After-Action Summary: ko serve HTTP daemon

## What Was Done

Implemented a minimal HTTP server for programmatic access to ko commands via `ko serve`.

### Files Added
- **cmd_serve.go** (145 lines): HTTP server with POST /ko endpoint, whitelist enforcement, graceful shutdown
- **cmd_serve_test.go** (236 lines): Unit tests for handler validation, JSON parsing, whitelist security
- **testdata/serve/basic.txtar**: Integration test verifying command wiring and flag parsing
- **specs/http_server.feature**: Behavioral specification (added during review to satisfy INVARIANTS.md)

### Files Modified
- **main.go**: Added `case "serve": return cmdServe(rest)` to run() switch and help text
- **ko_test.go**: Added TestServe function following existing pattern

## Implementation Decisions

### 1. Whitelist as map[string]bool
Used a map instead of slice for O(1) lookup. The 19 whitelisted commands match the ticket exactly: ls, ready, blocked, resolved, closed, query, show, questions, answer, close, reopen, block, start, bump, note, status, dep, undep, agent.

Explicitly excluded `create` and `add` per plan's open question #1 — ticket creation via HTTP should be a separate security decision.

### 2. Error response format
Success (exit 0): `200 OK`, `Content-Type: text/plain`, body is stdout
Failure (exit non-zero or validation): `400 Bad Request`, `Content-Type: application/json`, body is `{"error": "..."}`

This diverges from the ticket's "stderr → ignored (or logged)" in that stderr becomes part of the error response body when the command fails. The alternative (discarding stderr) would make debugging impossible for API consumers.

### 3. CombinedOutput vs separate stdout/stderr
Used `cmd.CombinedOutput()` which merges stdout and stderr. This simplifies the handler (no need to separately capture streams) and matches the ticket's intent: return "stdout" on success, return error context on failure.

### 4. Graceful shutdown timeout
Hardcoded 5-second timeout for graceful shutdown. This is not configurable but matches standard practice (Docker's default is 10s, systemd is 90s; 5s is conservative for a single-request HTTP server).

### 5. Integration test limitations
The testscript integration test validates command registration and flag parsing but cannot fully test the HTTP server (testscript doesn't support long-running background processes with concurrent curl requests). Full HTTP functionality is tested in Go unit tests using httptest instead.

### 6. Server startup synchronization
Used a channel (`serverStarted`) to ensure the goroutine prints the listening message before cmdServe blocks on signal handling. This prevents race conditions in tests or scripts that need to know when the server is ready.

## Security Considerations

- **Whitelist enforcement**: Only 19 known-safe ko subcommands allowed. Shell commands (sh, bash), filesystem operations (rm, mv, cp), and eval/exec are blocked.
- **No shell invocation**: Uses `exec.Command(os.Args[0], argv...)` directly, not via shell. No opportunity for command injection through argv.
- **No auth**: Ticket specifies "no auth for now (single-user, localhost)". This is documented as a deliberate choice; adding auth should be a separate ticket.
- **No CORS/TLS**: Per ticket requirements, these are explicitly excluded.

## Testing

- Unit tests: 2 test functions, 9 scenarios covering method validation, JSON parsing, whitelist enforcement, dangerous command rejection
- Integration test: Verifies command registration in main.go and flag parsing
- Manual verification: Server startup, valid/invalid subcommands, graceful shutdown

All tests pass. Build succeeds. Both new files under 500-line limit.

## Compliance with INVARIANTS.md

- ✅ Zero external runtime dependencies (net/http is stdlib)
- ✅ Files under 500 lines (145 and 236)
- ✅ Errors to stderr with non-zero exit codes
- ✅ CLI error messages human-readable
- ✅ Tests added (unit + integration)
- ✅ **Spec added** (specs/http_server.feature — created during review stage)

The spec was missing from the implementation stage output. I added it during review to satisfy the "every behavior has a spec" invariant. Future implementations should include the spec alongside the code.

## Notes for Future Work

1. **Authentication**: Current implementation has no auth. Before exposing this server beyond localhost, add authentication (API keys, OAuth, mutual TLS, etc.).

2. **Rate limiting**: No request throttling. A malicious or buggy client could spawn unbounded ko processes.

3. **Concurrency**: Each request spawns a new ko process. Under high load, consider a worker pool or request queue.

4. **Logging**: Server logs to stdout/stderr via http.Server defaults. For production use, consider structured logging (JSON) with request IDs for traceability.

5. **Content-Type validation**: Handler accepts any Content-Type and relies on JSON unmarshaling to fail. Could add explicit `application/json` check if needed.

6. **Timeout per request**: No timeout on command execution. A long-running `ko agent` command could block indefinitely. Consider adding exec timeout context.

7. **Health endpoint**: No `/health` or `/ping` endpoint for monitoring. Could add if needed for orchestration/load balancers.

8. **Metrics**: No Prometheus/statsd metrics. Could add counters for requests, errors, command duration if observability is needed.
