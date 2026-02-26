Looking at this ticket, I can classify it clearly:

**The ticket describes:** An unwanted behavior where the agent leaves uncommitted code changes in the working tree when a build fails. This creates confusion and prevents clean subsequent builds.

**The ticket specifies:** A clear fix strategy—the `on_fail` hook should stash or reset the worktree, with stash being the recommended approach.

**Classification:** This is a **task**. The ticket describes what to build (clean worktree handling on build failure) with a clear solution path (implement stashing in the on_fail hook).

```json
{"disposition": "route", "workflow": "task"}
```
