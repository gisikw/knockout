## Goal
Add an SSE endpoint to `ko serve` that broadcasts real-time ticket updates by tailing the global events.jsonl file.

## Context

The existing `ko serve` implementation (cmd_serve.go) provides a `/ko` POST endpoint that proxies whitelisted ko commands. The server uses graceful shutdown with signal handling.

The global mutation event stream lives at `~/.local/state/knockout/events.jsonl` (via `mutationEventPath()` in mutation.go). Each event is a JSONL line with:
- `ts` (timestamp)
- `event` (e.g., "status", "create", "bump")
- `project` (absolute path to project root)
- `ticket` (ticket ID)
- `data` (optional mutation details)

The reference implementation in punchlist-server/events.go demonstrates:
- Tail-based event streaming with fsnotify-style polling (200ms sleep on EOF)
- Handling of partial lines across reads
- Seek to end on open (only new events)
- Retry logic for file-not-found and errors
- Graceful degradation (best-effort)

The `ko query` command (cmd_query.go) outputs all tickets as JSONL using `resolveProjectTicketsDir()` to support both local projects and cross-project queries via `#tag` syntax.

Project paths are mapped to project root via `ProjectRoot(ticketsDir)` which returns the absolute path of `.ko/tickets/` minus two levels (ticket.go:501).

## Approach

Add a new GET `/subscribe/{project}` endpoint to cmd_serve.go's mux. The handler will:
1. Resolve the project parameter (interpret `#tag` syntax or treat as local relative path)
2. Send initial snapshot via `ko query --json`
3. Launch a goroutine to tail events.jsonl and filter by matching project path
4. On each matching mutation, re-run `ko query --json` and send as SSE event
5. Use monotonic counter for SSE `id:` field
6. Clean up on client disconnect

The tailing will use the punchlist-server pattern: open file, seek to end, read loop with EOF retry, parse JSONL lines, filter by project match.

Multiple subscribers will share a single global tailer (singleton pattern with mutex-protected subscriber list and broadcast). The tailer launches on first subscriber and persists until server shutdown.

## Tasks

1. [cmd_serve.go:handleSubscribe] — Add `handleSubscribe` function that:
   - Parses project path parameter from URL
   - Resolves it to absolute project root (handle `#tag` syntax via registry)
   - Executes `ko query --json` for initial snapshot
   - Sends SSE headers (`Content-Type: text/event-stream`, `Cache-Control: no-cache`, etc.)
   - Sends `retry: 3000` directive
   - Sends initial snapshot as event with id=0
   - Registers this subscriber with the global event tailer
   - Blocks on client disconnect or subscriber channel close
   - Deregisters subscriber on exit
   Verify: curl to endpoint returns SSE headers and initial data.

2. [cmd_serve.go:eventTailer] — Add global tailer singleton with:
   - `subscribers` map (project path -> list of subscriber channels)
   - `mu sync.Mutex` for thread-safe subscriber management
   - `started bool` flag to ensure single tailer goroutine
   - Methods: `Subscribe(project string, ch chan<- string)`, `Unsubscribe(project, ch)`, `Start()`
   Verify: multiple subscribers can register/unregister without races.

3. [cmd_serve.go:tailEventStream] — Implement tail goroutine using punchlist-server pattern:
   - Open `mutationEventPath()` (from mutation.go)
   - Seek to end (io.SeekEnd)
   - Read loop with 200ms sleep on EOF
   - Parse JSONL lines, unmarshal to MutationEvent
   - Filter by project path, broadcast to matching subscribers
   - Handle file-not-found with retry (wait for file creation)
   Verify: mutations written to events.jsonl trigger broadcasts.

4. [cmd_serve.go:broadcastToProject] — Add function to re-query and send SSE:
   - Execute `ko query --json {project}` via internal command dispatch
   - Format as SSE event: `id: {counter}\ndata: {json}\n\n`
   - Send to all subscribers for this project
   - Increment global event counter
   Verify: SSE events have monotonic IDs and valid format.

5. [cmd_serve.go:cmdServe] — Wire handleSubscribe into mux:
   - Add `mux.HandleFunc("/subscribe/", handleSubscribe)` after existing `/ko` handler
   - Launch global tailer on server start: `eventTailer.Start()`
   Verify: `go test ./...` passes, server starts without errors.

6. [cmd_serve_test.go:TestSubscribeHandler] — Add test for SSE endpoint:
   - Create temp project with tickets
   - Start server with test mux
   - Connect to `/subscribe/{project}`, verify SSE headers
   - Verify initial snapshot received
   - Write mutation event to mock events.jsonl
   - Verify follow-up SSE event received
   Verify: test passes in isolation and with full suite.

7. [cmd_serve_test.go:TestSubscribeMultiple] — Add test for concurrent subscribers:
   - Multiple goroutines connect to same project
   - Emit single mutation event
   - Verify all subscribers receive the event
   Verify: no race conditions, all events delivered.

## Open Questions

1. **Project path resolution strategy**: The endpoint receives `{project}` as a URL path segment. Should we:
   - Allow arbitrary filesystem paths (security concern: directory traversal)
   - Only support `#tag` syntax from the registry (requires registry for all projects)
   - Support both but validate/sanitize filesystem paths

   **Recommendation**: Support both. Use `#tag` syntax for registry lookup. For non-`#` values, treat as relative path from ko serve working directory, resolve to absolute path, validate it contains `.ko/tickets/`, then use `ProjectRoot()`. Reject paths with `..` or outside server CWD.

2. **Event stream backpressure**: If a subscriber's channel blocks (slow client), should we:
   - Drop events (best-effort delivery)
   - Buffer with a limit then drop
   - Block the tailer (affects all subscribers)

   **Recommendation**: Use buffered channels (buffer size 10) and non-blocking send. If channel is full, drop the event. SSE clients can handle missed events via full re-sync on reconnect.

3. **Tailer lifecycle**: Should the tailer goroutine:
   - Run continuously once started (simpler, uses one goroutine)
   - Shut down when no subscribers remain (more efficient)

   **Recommendation**: Run continuously. The overhead is negligible (one goroutine sleeping on EOF), and it simplifies subscriber management (no start/stop coordination).

4. **Testing event emission**: Tests need to write to events.jsonl, but the real file is global state. Should we:
   - Mock `mutationEventPath()` with a test-specific path
   - Use a real temp file and override XDG_STATE_HOME
   - Skip integration test, only unit test components

   **Recommendation**: Override `XDG_STATE_HOME` in tests to use a temp directory. This tests the real code path without polluting the global state.
