---
id: ko-f7d1
status: blocked
deps: []
created: 2026-02-17T03:43:04Z
type: task
priority: 2
---
# Retry transitions and build failures should emit events to the event log

## Problem

When a build step fails and ko retries it, or when a step is killed (e.g.,
hung process), the retry/failure transition is invisible in the event log.
The log shows `node_start` for the step, then nothing until the retry's
`node_start` appears — or worse, nothing at all if the failure killed the
build. This makes it impossible to diagnose what happened from the log alone.

Observed: a 6-hour hung implement step was killed externally. The event log
showed `node_start` for implement with no corresponding complete/fail event.
The loop retried successfully, but the retry was also invisible in the log.

## Expected events

- `node_fail` — when a step fails (with reason: timeout, process killed,
  non-zero exit, etc.)
- `node_retry` — when a failed step is being retried (with attempt number)
- `build_error` — when the build itself fails (on_succeed hook failure, etc.)

## Acceptance Criteria

- [ ] Step failures emit a `node_fail` event with the error reason
- [ ] Retries emit a `node_retry` event with the attempt number
- [ ] Build-level errors (hook failures, etc.) emit a `build_error` event
- [ ] The event log is a complete record — no silent gaps between events

## Notes

**2026-02-24 13:07:18 UTC:** ko: FAIL at node 'classify' — node 'classify' failed after 3 attempts: disposition missing required 'disposition' field

**2026-02-24 14:52:29 UTC:** ko: FAIL at node 'assess' — node 'assess' failed after 3 attempts: 'route' disposition missing required 'workflow' field
