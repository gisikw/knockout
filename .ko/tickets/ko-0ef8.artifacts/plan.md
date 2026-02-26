## Goal
Clean worktree on build failure by adding git stash to on_fail hooks.

## Context
When a build fails at any node (e.g., verify test failure), the agent leaves uncommitted code changes in the working tree. This creates two problems:

1. **Subsequent builds start dirty**: The next ticket's build starts with leftover changes from the previous failed build, making it unclear what code belongs to which ticket.
2. **Provenance confusion**: When humans inspect failed tickets, uncommitted changes could be from the current ticket or from a previous failed build.

The build system already has an `on_fail` hook mechanism (build.go:106-109, 116-119) that runs shell commands when builds fail. The hooks receive environment variables including `$TICKET_ID` and run from the project root (build.go:654-688).

Current dogfood pipeline (.ko/pipeline.yml:69-70) has:
```yaml
on_fail:
  - git add .ko/tickets/ && git diff --cached --quiet || git commit -m "ko: ${TICKET_ID} (blocked)"
```

This commits ticket metadata changes but doesn't clean the worktree. The ticket suggests `git stash` with the ticket ID in the message to preserve failed work for potential retry while preventing it from polluting subsequent builds.

Key files:
- build.go:654-688 — `runHooks` function executes hook commands
- build.go:106-109, 116-119 — on_fail hooks invoked after build failure
- .ko/pipeline.yml:69-70 — current on_fail hook configuration
- INVARIANTS.md:100-102 — documents that on_close runs after ticket is closed (safe for destructive operations)

## Approach
Add a git stash command to the on_fail hooks in the dogfood pipeline. The stash message will include the ticket ID for traceability. This will:

1. Preserve uncommitted changes for manual inspection/retry if needed
2. Return the worktree to a clean state before the next build starts
3. Follow the pattern of existing hooks (uses $TICKET_ID env var)

The stash should happen after the ticket metadata commit, so we preserve both the ticket state and the code changes separately. We'll use `git stash push -u -m` to include untracked files and add a descriptive message.

## Tasks
1. [.ko/pipeline.yml:69-71] — Add git stash command to on_fail hook list
   - Add after the existing ticket commit command
   - Use: `git stash push -u -m "ko: ${TICKET_ID} failed build"`
   - The `-u` flag includes untracked files (new files created during build)
   - The message makes it easy to identify which ticket's changes are in the stash
   Verify: No test changes needed; this is a configuration change.

## Open Questions
None — this is a straightforward configuration change using existing hook infrastructure. The stash approach preserves work (allowing manual recovery if needed) while ensuring clean builds going forward.
