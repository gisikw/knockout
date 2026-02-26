## Summary

**Task 1: Add git stash command to on_fail hook list** ✓

Modified `.ko/pipeline.yml:69-71` to add a git stash command after the existing ticket metadata commit command.

**Changes:**
- Added `git stash push -u -m "ko: ${TICKET_ID} failed build"` as the second command in the `on_fail` hook list
- The `-u` flag includes untracked files (new files created during the build)
- The stash message includes the ticket ID for easy identification
- This runs after the ticket metadata commit, preserving both ticket state and code changes separately

**Verification:**
As noted in the plan, no test changes are needed since this is a configuration change. The on_fail hook will now:
1. First commit any ticket metadata changes to preserve ticket state
2. Then stash all uncommitted changes (including untracked files) with a descriptive message

This ensures the next build starts with a clean worktree while preserving failed build artifacts for potential manual inspection or retry.

**No deviations from the plan.**
