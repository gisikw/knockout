The plan is complete. I've investigated the codebase and written a detailed implementation plan that:

## Summary

The plan implements pipeline support for the `resolved` status through two mechanisms:

1. **Workflow-level configuration**: Add `on_success: resolved|closed` field to workflows in pipeline.yml, allowing workflows like `research` and `bug` to land tickets in `resolved` status for human review instead of automatically closing them.

2. **Disposition override**: Add a new `resolved` disposition type that decision nodes can emit to explicitly land a ticket in `resolved` status, overriding the workflow default.

The implementation:
- Adds the `OnSuccess` field to the `Workflow` struct
- Updates the YAML parser to read this config
- Threads the final workflow name through `runWorkflow()` to determine which workflow's config to use
- Applies the appropriate status transition in `RunBuild()` based on the workflow config
- Supports both the workflow-level default and disposition-level override
- Maintains backward compatibility (defaults to `closed`)

All changes are localized to the build pipeline and don't require modifications to the ticket status system (since `resolved` was already added in ko-dd51).
