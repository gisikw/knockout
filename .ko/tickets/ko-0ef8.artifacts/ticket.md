---
id: ko-0ef8
status: open
deps: []
created: 2026-02-26T01:24:41Z
type: task
priority: 1
---
# Agent should not leave dirty worktree on build failure

When a build fails at verify (or any node), the agent leaves uncommitted code changes in the working tree. This prevents clean builds of subsequent tickets and creates confusion about provenance when humans intervene. The on_fail hook should stash or reset the worktree so the next build starts clean. Options: git stash with ticket ID in message, git checkout -- ., or commit to a branch. Stash is probably best — preserves the work for retry without polluting main.
