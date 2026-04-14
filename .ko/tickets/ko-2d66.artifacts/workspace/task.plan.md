knockout — just --list for recipes
The plan is written to `.ko/tickets/ko-2d66.artifacts/plan.md`. Here's a summary:

**Root cause:** Double-fork daemonization reparents the agent to `init` but can't escape the invoking one-shot's cgroup. systemd v2 kills the whole cgroup on service exit.

**Fix:** Three pieces:

1. **`cmd_serve.go`** — Add `serveAddrPath()` + write `XDG_STATE_HOME/knockout/serve.addr` when the server starts, delete it on shutdown.

2. **`cmd_agent.go`** — Add `tryDelegateAgentStart()` that reads the addr file and POSTs `["agent", "_daemonize", ...]` to `ko serve`'s existing `/ko` endpoint. Since `_daemonize` exits immediately after spawning the grandchild, `CombinedOutput()` returns quickly with the grandchild PID — and that grandchild is now in `ko serve`'s cgroup. Modify `cmdAgentStart` to try this first, fall back to the current double-fork if `ko serve` isn't reachable.

3. **`cmd_agent_test.go` + `specs/agent_lifecycle.feature`** — Unit test for the delegation function (mock HTTP server), and a new spec file covering the four observable behaviors from the ticket obligations.

**Key trade-off noted:** Agent now dies if `ko serve` restarts (rather than surviving everything). Acceptable per the ticket's intent — `ko serve` is the designated long-lived host.
