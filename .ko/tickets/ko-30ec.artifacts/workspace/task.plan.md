Perfect! I've created the implementation plan. The plan is now ready for review at `/home/dev/Projects/knockout/.ko/tickets/ko-30ec.artifacts/plan.md`.

## Summary

I've investigated the codebase and written a plan for adding project-scoped execution to the POST /ko endpoint. Key findings:

1. **Existing Pattern**: The `/subscribe` endpoint already implements project resolution (lines 220-254 of cmd_serve.go) using registry lookup for #tags and absolute paths
2. **Working Directory Pattern**: The codebase sets `cmd.Dir` on exec.Cmd instances (seen in build.go and cmd_agent.go)
3. **Registry Module**: Provides all needed functionality for #tag → path resolution

The plan follows these established patterns to add an optional `"project"` field to the JSON request body, resolve it the same way the subscribe endpoint does, and set the command's working directory accordingly. When the field is absent, it preserves current behavior (uses server's cwd).
