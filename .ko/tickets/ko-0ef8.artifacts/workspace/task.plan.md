Perfect! I've written the plan to the artifact directory. The plan is complete and ready for review.

## Summary

I've investigated the codebase and created a plan to address the dirty worktree issue on build failures. Here's what I found:

**The Problem**: When builds fail at any node (like verify), uncommitted code changes remain in the worktree, causing subsequent builds to start dirty and creating confusion about which changes belong to which ticket.

**The Solution**: Add a `git stash` command to the `on_fail` hooks in `.ko/pipeline.yml`. This will:
- Preserve failed work for potential manual inspection/retry
- Clean the worktree before the next build starts
- Use the ticket ID in the stash message for traceability

The plan leverages the existing hook infrastructure that's already in place — no code changes needed, just a one-line configuration update to the pipeline.
