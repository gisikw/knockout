## Goal
Append a JSONL event to per-ticket history whenever the ticket is touched (status/note/dep changes) or when a build step transition occurs.

## Context
The codebase already has three event logging systems:

1. **EventLogger** (eventlog.go) — Per-build, truncated log written to `$KO_EVENT_LOG` if set. Logs workflow/node events during a single build run.

2. **BuildHistoryLogger** (buildhistory.go) — Per-ticket append-only JSONL at `.ko/tickets/<id>.jsonl`. Currently logs build lifecycle events: `build_start`, `build_complete`, `workflow_start`, `node_start`, `node_complete`. Opened in build.go:79, closed at end of RunBuild.

3. **MutationEvent** (mutation.go) — Global append-only JSONL at `~/.local/state/knockout/events.jsonl`. Logs ticket mutations: `create`, `status`, `note`, `dep`, `undep`, `bump`. Emitted from cmd_*.go files and build.go:530 (setStatus).

The ticket title asks for "JSONL append on every ticket touch and build step transition". Currently:
- Ticket touches (status, note, dep, bump) emit to the **global** mutation log via EmitMutationEvent
- Build step transitions emit to the **per-ticket** build history via BuildHistoryLogger

**Gap:** Ticket mutation events (status, note, dep, etc.) are **not** being written to the per-ticket JSONL history. Only build-related events go there.

## Approach
Add per-ticket mutation event logging to BuildHistoryLogger. Whenever EmitMutationEvent is called (which already fires on every ticket touch), also append the same event to the per-ticket `.ko/tickets/<id>.jsonl` file. This unifies both build events and mutation events into a single per-ticket audit trail.

## Tasks
1. [buildhistory.go:BuildHistoryLogger] — Add a `Mutation(event string, data map[string]interface{})` method that writes mutation events in the same format as EmitMutationEvent.
   Verify: Method compiles, follows existing emit pattern.

2. [mutation.go:EmitMutationEvent] — After writing to the global log, open the per-ticket BuildHistoryLogger and append the same event using the new Mutation method. Handle errors silently (best-effort, same as current behavior).
   Verify: Mutation events appear in both global and per-ticket JSONL files.

3. [buildhistory.go:BuildHistoryLogger] — Ensure the per-ticket JSONL file can be opened/appended even when not in an active build context (mutation.go is called from cmd_*.go outside of builds).
   Verify: `ko note <id> <text>` and `ko status <id> open` both append to `.ko/tickets/<id>.jsonl` successfully.

## Open Questions
None — the pattern is clear. Mutation events should mirror to both the global registry and the per-ticket history. The per-ticket log becomes the canonical audit trail for all ticket lifecycle events.
