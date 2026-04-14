## Goal

`ko agent start` delegates agent spawning to `ko serve` via HTTP so the agent lives in `ko serve`'s cgroup and survives one-shot systemd callers.

## Context

**The problem:** `cmdAgentStart` (`cmd_agent.go:133`) currently double-forks via `ko agent _daemonize`, which is supposed to reparent the agent to `init`. This works for process-tree checks but fails under systemd cgroup v2: when the invoking one-shot service exits, systemd kills the entire cgroup regardless of reparenting.

**The existing infrastructure:**
- `ko serve` (`cmd_serve.go`) runs an HTTP server on a configurable port (default 19876) with a `/ko` endpoint (`POST {"argv": [...]}`) that executes `ko` subcommands via `exec.Command(...).CombinedOutput()`.
- `"agent"` is already in the `/ko` whitelist, so `["agent", "_daemonize", ...]` is an allowed call.
- `_daemonize` (`cmd_agent.go:218`) spawns its grandchild and exits immediately — `CombinedOutput()` returns quickly. The grandchild is in `ko serve`'s cgroup because it's forked from a child of `ko serve`.
- Global state lives in `XDG_STATE_HOME/knockout/` (see `mutationEventPath()` in `mutation.go`).

**The fix:** When `ko serve` starts, it writes its listening address to `XDG_STATE_HOME/knockout/serve.addr`. `ko agent start` reads this file and POSTs to `/ko` with `_daemonize` args, getting back the grandchild PID. If the file is absent or the connection fails, it falls back to the existing double-fork.

**Cgroup semantics:** The agent loop's cgroup membership is inherited from `_daemonize`, which is inherited from `ko serve`'s HTTP handler. When `_daemonize` exits, the agent is reparented to init but stays in `ko serve`'s cgroup. SIGTERM from `ko agent stop` (sent by PID) still reaches the process regardless of cgroup. ✓

**Trade-off:** The agent now dies if `ko serve` itself restarts. This is acceptable — `ko serve` is a long-lived service and the ticket explicitly identifies it as the intended host cgroup.

**Line count warning:** `cmd_agent.go` is 634 lines (exceeds 500-line invariant). Don't make it bigger. The two new functions (`serveAddrPath` and `tryDelegateAgentStart`) each stay under 25 lines. If a split is needed, that's a separate ticket.

## Approach

Add `serveAddrPath()` to `cmd_serve.go` alongside `cmdServe`, and write/delete the addr file during server lifecycle. In `cmd_agent.go`, add `tryDelegateAgentStart()` that reads the addr file and POSTs to `/ko` with `_daemonize` args; modify `cmdAgentStart` to call it first and fall back to double-fork on failure. Add a spec for the delegation behavior and a unit test for the delegation function.

## Tasks

1. **[cmd_serve.go:serveAddrPath]** — Add `serveAddrPath() string` that returns `XDG_STATE_HOME/knockout/serve.addr` (mirrors `mutationEventPath()` in `mutation.go`). In `cmdServe`, after the `serverStarted` channel is closed, write the listening address to this file. In the shutdown path (after `server.Shutdown`), delete the file. Use `os.Remove` (best-effort, no error check on delete).
   Verify: `go build ./...` passes. Manual: `ko serve &`, check that `serve.addr` exists and contains the port; kill serve, check file is gone.

2. **[cmd_agent.go:tryDelegateAgentStart]** — Add `tryDelegateAgentStart(ticketsDir string, extraArgs []string) (int, error)` that:
   - Reads serve addr from `serveAddrPath()`. Returns an error if file is missing.
   - Builds the `_daemonize` argv: `["agent", "_daemonize", logPath, projectRoot, self, "agent", "loop"] + extraArgs`.
   - POSTs `{"argv": [...]}` to `http://<addr>/ko` with a 5-second timeout.
   - On non-200 response, returns error with the response body.
   - On success, parses the response body as a PID (trims whitespace, `strconv.Atoi`).
   - Returns the PID on success.
   Verify: unit test passes (Task 4).

3. **[cmd_agent.go:cmdAgentStart]** — At the top of `cmdAgentStart`, after the existing-agent checks, call `tryDelegateAgentStart(ticketsDir, args)`. On success: write the returned PID to the pid file, print the success message, and return 0. On failure: log the failure reason to stderr at debug level (e.g. `"ko agent start: ko serve not available, using double-fork (%v)\n"`) and fall through to the existing double-fork logic.
   Verify: `go build ./...` passes. Behavior unchanged when `ko serve` is not running.

4. **[cmd_agent_test.go]** — Add `TestTryDelegateAgentStart` with two sub-cases:
   - **serve not running**: addr file absent → returns error, no HTTP call made.
   - **serve running (mock)**: spin up an `httptest.NewServer` that expects a POST to `/ko` with `_daemonize` in argv and returns `"12345\n"`; write the server's addr to a temp `serve.addr` file overriding `XDG_STATE_HOME`; call `tryDelegateAgentStart`; assert returns PID 12345 and no error.
   Verify: `go test ./... -run TestTryDelegateAgentStart` passes.

5. **[specs/agent_lifecycle.feature]** — Create new feature file with scenarios covering the delegation obligations from the ticket:
   - Scenario: "agent start delegates to ko serve when available" — ko serve running → agent appears as child of ko serve.
   - Scenario: "agent survives caller exit when spawned via ko serve" — agent is in ko serve's cgroup.
   - Scenario: "agent start falls back to double-fork when ko serve is unavailable" — ko serve not running → starts via double-fork.
   - Scenario: "agent stop still terminates the agent after serve-delegated start" — PID file written correctly, stop sends SIGTERM.
   Verify: file exists and is consistent with implementation.

## Open Questions

1. **Addr file race on serve restart**: If `ko serve` restarts quickly, the old addr file might briefly point to nothing. The 5-second timeout on the HTTP client handles this (connection refused returns immediately). No action needed, but worth noting.

2. **`maybeAutoAgent` interaction**: `maybeAutoAgent` calls `ko agent start` via subprocess. With delegation, `ko agent start` will try to delegate. If `ko serve` is the process that triggered `maybeAutoAgent` (e.g., via a hook), and `ko serve` calls `ko agent start` which calls back into `ko serve`... this is fine — it's a synchronous HTTP call from a subprocess, not a re-entrant goroutine call. No deadlock risk. Confirmed no action needed.

3. **Port written to addr file**: Write just the port number (e.g., `"19876"`) or the full address (e.g., `"127.0.0.1:19876"`)? Recommend writing the full `net.JoinHostPort("127.0.0.1", *port)` form so the reader doesn't need to reconstruct it.
