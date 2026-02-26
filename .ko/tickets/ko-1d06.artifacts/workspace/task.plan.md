I've created a comprehensive plan for adding post-loop hook support to the knockout pipeline system. The plan adds an `on_loop_complete` configuration option that runs shell commands once after the entire agent loop finishes (as opposed to the existing hooks that run per-ticket).

Key design decisions:
- **New hook field**: `OnLoopComplete` in the Pipeline struct
- **Environment variables**: Loop summary stats (processed, succeeded, failed, etc.) available to hooks
- **Execution timing**: Runs after loop completion, before cleanup, regardless of stop reason
- **Error handling**: Hook failures are logged but don't affect loop exit code
- **Follows existing patterns**: Mirrors the architecture of `on_succeed`/`on_fail`/`on_close` hooks

The implementation is straightforward, following established patterns in the codebase for YAML parsing, hook execution, and testing. All tasks have clear verification steps and the plan includes spec, test, documentation, and example updates per the INVARIANTS.md requirements.
