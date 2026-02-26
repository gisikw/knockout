Perfect! I've investigated the codebase and created a plan for implementing the new SSE endpoint. Here's a summary of what I found and planned:

## Investigation Summary

The codebase already has a solid SSE foundation in `cmd_serve.go`:
- A `tailer` struct that manages event broadcasting to subscribers
- An existing `/subscribe/` endpoint that streams ticket updates
- Mutation events that trigger broadcasts when tickets change
- Agent status tracking via PID files and lock files in `cmd_agent.go`

## Plan Overview

I've written a plan to add a new `/status/` SSE endpoint that streams both tickets and agent status. The key decisions in the plan:

1. **New endpoint** (`/status/`) rather than modifying the existing `/subscribe/` for backward compatibility
2. **Format**: Agent status and tickets will be sent as typed JSON objects (`{"type":"agent",...}` and `{"type":"ticket",...}`)
3. **Polling**: The tailer will poll agent status every 2 seconds to detect changes
4. **Implementation**: Reuse the existing `agentStatusJSON` struct from `cmd_agent.go`

## Open Questions

The plan includes three open questions about:
- SSE event format (resolved: use type discriminators within data lines)
- Polling frequency (resolved: 2 seconds)
- Backward compatibility (resolved: keep `/subscribe/` unchanged, add new endpoint)

The plan is ready for review at `$KO_ARTIFACT_DIR/plan.md`.
