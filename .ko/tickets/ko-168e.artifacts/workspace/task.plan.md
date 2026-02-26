The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-168e.artifacts/workspace/plan.md`.

## Summary

I've investigated the codebase and created an implementation plan for removing the `project` field from the POST /ko endpoint payload. 

**Key findings:**
- The serve endpoint currently accepts a special `project` field in the JSON payload and handles project resolution server-side
- Dependency ticket ko-6f77 added --project flag support to commands, making the special payload field redundant
- The ticket author confirmed the approach should be "Remove payload"

**The plan involves:**
1. Removing the `Project` field from the request struct
2. Deleting the project resolution logic (lines 385-421) that set cmd.Dir
3. Updating tests to pass --project in argv instead of as a payload field
4. Simplifying TestServeProjectScoped to test flag passthrough rather than server-side resolution

This is a straightforward simplification that removes a special case from the serve API. Commands already handle --project internally via `resolveProjectTicketsDir()`, so the server can be a simple passthrough.
