# Summary: Clean Worktree on Build Failure

## What Was Done

Added a `git stash` command to the `on_fail` hook in `.ko/pipeline.yml` to automatically clean the worktree when builds fail. This prevents subsequent builds from starting with leftover uncommitted changes from previous failed builds.

**Change:** Added line 71 to `.ko/pipeline.yml`:
```yaml
on_fail:
  - git add .ko/tickets/ && git diff --cached --quiet || git commit -m "ko: ${TICKET_ID} (blocked)"
  - git stash push -u -m "ko: ${TICKET_ID} failed build"
```

## Implementation Decisions

- **Stash approach chosen:** Preserves failed build artifacts for manual inspection/retry while ensuring clean subsequent builds
- **Placement:** Runs after ticket metadata commit so both ticket state and code changes are preserved separately
- **Flags used:**
  - `-u`: Includes untracked files (new files created during build)
  - `-m`: Adds descriptive message with ticket ID for easy identification

## Future Reader Notes

- This is a dogfood pipeline configuration change, not a code change
- No test updates needed (existing `testdata/pipeline/build_on_fail_hook.txtar` validates hook execution)
- The stash can be recovered manually using `git stash list` to find the ticket ID and `git stash apply stash@{N}`
- If stash behavior becomes problematic, alternatives include `git reset --hard` (destructive) or committing to a ticket-specific branch

## Outcome

All tests pass. The on_fail hook will now leave a clean worktree after build failures while preserving failed changes in the stash for potential manual recovery.
