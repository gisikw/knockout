---
id: ko-ou50
status: closed
deps: []
links: []
created: 2026-02-16T00:07:41Z
type: task
priority: 2
assignee: Kevin Gisi
---
# Quiet flag: suppress stdout, summary-only on exit

Add `--quiet` flag to `ko build` and `ko loop` that suppresses all stdout
output during execution. On exit, emit a single summary line.

**Behavior:**
- `ko build --quiet <id>` → no stdout during build, then:
  `"build: exo-avqk SUCCEED. See $KO_EVENT_LOG for details"`
- `ko loop --quiet` → no stdout during loop, then:
  `"loop complete: 7 processed (5 succeeded, 0 failed, 2 blocked, 0 decomposed). See $KO_EVENT_LOG for details"`
- If `KO_EVENT_LOG` is not set, omit the "See..." part
- Stderr unchanged (errors still go to stderr)

**Motivation:**
When ko is invoked as a tool call from Claude Code, the entire stdout output
becomes the tool response. A multi-hour loop could produce thousands of lines,
overwhelming the context window. With `--quiet`, the caller gets a clean
summary and real-time progress goes through the event log side channel.

Depends on: ko-38xt (KO_EVENT_LOG)
