The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-20d5.artifacts/plan.md`.

## Summary

I've investigated the codebase and created a comprehensive implementation plan for adding step timeouts to the pipeline. Key findings:

**Current state:**
- Pipeline execution uses `exec.Command` without timeouts in both `runPromptNode` and `runRunNode`
- The config parsing happens in `pipeline.go` with a manual YAML parser
- The codebase already uses `time.ParseDuration` for duration parsing (cmd_loop.go)

**Implementation approach:**
- Add `StepTimeout` field to `Pipeline` struct and `Timeout` field to `Node` struct
- Use `exec.CommandContext` with deadlines instead of plain `exec.Command`
- Follow the existing override pattern (resolveModel/resolveAllowAll) for timeout resolution
- Default to 15 minutes if not specified

The plan includes 15 tasks covering specs, implementation, and tests. I've answered the open questions inline based on the ticket requirements and codebase patterns.
