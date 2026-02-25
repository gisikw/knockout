---
id: ko-759a
status: open
deps: [ko-8e81]
created: 2026-02-25T05:49:50Z
type: task
priority: 2
---
# ko serve: SSE subscribe endpoint for real-time item updates

Add an SSE (Server-Sent Events) endpoint to ko serve:

    GET /subscribe/{project}

Implementation:
- On connect, immediately send full item list (ko query --json for the project)
- Tail the global events JSONL file (~/.local/state/knockout/events.jsonl)
  for mutations matching this project
- On each matching mutation, re-query and push full item list as an SSE event
- Use standard SSE format: `data: {...}\n\n`
- Include `id:` field on events (monotonic counter or timestamp) so clients
  can send Last-Event-ID on reconnect
- Include `retry: 3000` directive for client reconnection interval
- Multiple concurrent subscribers supported (fan-out from single tail)
- Clean up subscriber on disconnect

The event tailing pattern already exists in punchlist-server/events.go — reference
that implementation for the JSONL tail + fsnotify approach.

Depends on ko-8e81 (base ko serve command).
