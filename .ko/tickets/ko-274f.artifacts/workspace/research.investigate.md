## Summary

I've completed my investigation into how to programmatically determine if the working directory is clean before starting an agent. Here's what I found:

**Current State:** The codebase has no mechanism to check working directory cleanliness before agent startup. While there are `on_fail` hooks that clean up after builds fail (via `git stash`), nothing prevents starting with a dirty working directory.

**Key Findings:**

1. **`ko agent status` currently exposes:**
   - Whether agent is provisioned/running
   - PID and last log line
   - **Does NOT check git status**

2. **No pre-loop hooks exist** - only `on_succeed`, `on_fail`, and `on_close` hooks

**Recommended Solutions:**

1. **Extend `ko agent status --json`** (simpler, ~1 hour effort)
   - Add `working_dir_clean` boolean field
   - Check via `git status --porcelain`
   - Enables external scripts to check before starting agent

2. **Add `on_start` hooks to pipeline.yml** (more comprehensive, ~2-3 hours effort)
   - Run shell commands before loop begins
   - Users can enforce policies: fail-fast or auto-stash
   - Consistent with existing hook architecture

**Best approach:** Implement both (hybrid solution) for maximum flexibility:
- Status API for programmatic external checks
- `on_start` hooks for declarative enforcement in pipeline config

All findings, code references, and implementation details are documented in `/home/dev/Projects/knockout/.ko/tickets/ko-274f.artifacts/findings.md`.
