---
id: ko-e92f
status: closed
deps: [ko-76f7, ko-12c4]
links: []
created: 2026-02-15T16:35:05Z
type: feature
priority: 2
---
# Pipeline v2: workflow runner

Replace linear stage executor with workflow-aware runner. Tracks visit counts per node (max_visits, default 1). Handles route dispositions (jump to named workflow). Route-to-self restarts workflow from top. Terminates on fail/blocked/decompose/end-of-workflow. End of workflow = succeed.
