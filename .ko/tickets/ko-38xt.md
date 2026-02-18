---
id: ko-38xt
status: closed
deps: []
links: []
created: 2026-02-16T00:07:37Z
type: task
priority: 2
assignee: Kevin Gisi
---
# KO_EVENT_LOG: structured event log output

If the `KO_EVENT_LOG` environment variable is set to a file path, ko should
write structured JSONL events there during `build` and `loop` execution.

**Behavior:**
- Truncate the file on start (caller owns the path, ko owns it for the run)
- Write one JSON object per line for significant events
- Events: workflow_start, node_start, node_complete, workflow_complete, loop_summary
- Each event includes timestamp, ticket ID, and relevant metadata

**Event schema:**
```jsonl
{"ts":"2026-02-16T00:00:00Z","event":"workflow_start","ticket":"exo-avqk","workflow":"main"}
{"ts":"...","event":"node_start","ticket":"exo-avqk","workflow":"main","node":"triage"}
{"ts":"...","event":"node_complete","ticket":"exo-avqk","workflow":"main","node":"triage","result":"continue"}
{"ts":"...","event":"workflow_complete","ticket":"exo-avqk","outcome":"succeed"}
{"ts":"...","event":"loop_summary","processed":7,"succeeded":5,"failed":0,"blocked":2,"decomposed":0,"stop_reason":"empty"}
```

**Design notes:**
- If the env var is not set, no log file is written (zero behavior change)
- The env var contract: "you point me at a file, I own it for the duration of my run"
- Truncate-on-start is correct because watchers (inotify/fsnotify) track file handles,
  not paths — creating new files would break watchers
