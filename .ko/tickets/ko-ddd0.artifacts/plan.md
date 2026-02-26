## Goal
Add a new SSE endpoint that streams both ticket lists and agent status information to subscribers.

## Context
The codebase already has SSE infrastructure in `cmd_serve.go`:
- `tailer` struct manages the global event stream and broadcasts to subscribers
- `handleSubscribe` endpoint at `/subscribe/` currently streams only ticket lists
- Uses mutation events (written to `$XDG_STATE_HOME/knockout/events.jsonl`) to trigger broadcasts
- `ticketToJSON` helper in `cmd_list.go` serializes tickets for SSE data payloads
- Agent status is tracked via PID files and lock files (see `cmd_agent.go`)
- `cmdAgentStatus` returns structured JSON including `provisioned`, `running`, `pid`, and `last_log`

The existing SSE implementation:
1. Watches the mutation event file via `tailEventStream`
2. When a mutation occurs, re-queries tickets via `ListTickets`
3. Broadcasts JSONL ticket data to all subscribers for that project
4. Each SSE event has an incrementing `id:` field and multiple `data:` lines (one per ticket)

The new endpoint needs to stream both tickets AND agent status, likely requiring:
- Either a new endpoint (e.g., `/status/`) or enhancing `/subscribe/`
- Gathering agent status alongside tickets for the initial snapshot
- Broadcasting agent status changes when the agent starts/stops/updates
- Deciding on the SSE message format (separate events? combined data?)

## Approach
Add a new SSE endpoint `/status/` that streams both tickets and agent status. The initial snapshot will include both datasets, and broadcasts will be triggered by either ticket mutations or agent status changes (detected by polling the agent PID/log files).

The agent status will be sent as a separate SSE event interleaved with ticket events, using a distinct format (e.g., `data: {"type":"agent",...}` vs `data: {"type":"ticket",...}`).

## Tasks
1. [cmd_serve.go:tailer] — Add a background goroutine to the tailer that polls agent status periodically (every 2-3 seconds) and broadcasts status changes.
   Verify: Build succeeds, no compilation errors.

2. [cmd_serve.go] — Add a helper function `getAgentStatus(ticketsDir)` that returns the same `agentStatusJSON` struct used by `cmdAgentStatus`, for reuse in SSE broadcasts.
   Verify: Function compiles and returns correct status in both provisioned and not-provisioned cases.

3. [cmd_serve.go] — Create a new handler `handleStatusSubscribe` that serves SSE for both tickets and agent status. Pattern follows `handleSubscribe` but:
   - Initial snapshot sends agent status as first event (with a type discriminator in the JSON)
   - Ticket data follows, each line tagged with `"type":"ticket"`
   - Broadcasts include both ticket updates and agent status updates
   Verify: Handler compiles and serves SSE with correct headers.

4. [cmd_serve.go:cmdServe] — Register the new `/status/` route with the mux, pointing to `handleStatusSubscribe`.
   Verify: Server starts without errors, `/status/` endpoint is reachable.

5. [cmd_serve_test.go] — Add tests for the new `/status/` endpoint:
   - Test that initial snapshot includes agent status
   - Test that agent status broadcasts are received when status changes
   - Test that multiple subscribers receive both ticket and agent updates
   Verify: `go test ./...` passes, including new tests.

## Open Questions
**Format for agent status in SSE events:**
- Should agent status be sent as a separate SSE event (with its own `id:`), or inline as `data:` lines within the same event as tickets?
- Recommendation: Send agent status as the first `data:` line in each event, with `{"type":"agent",...}`, followed by ticket lines with `{"type":"ticket",...}`. This keeps events atomic (one event = full state snapshot) and simplifies client parsing.

**Polling frequency for agent status:**
- How often should the tailer poll agent status to detect changes?
- Recommendation: Poll every 2 seconds. This is fast enough to feel responsive but won't cause excessive I/O (just reading PID files and checking process existence).

**Backward compatibility:**
- Should the existing `/subscribe/` endpoint remain unchanged, or should we enhance it to include agent status?
- Recommendation: Keep `/subscribe/` unchanged (tickets only) and add `/status/` as a new endpoint that includes both. This allows clients to opt into the richer data without breaking existing integrations.
