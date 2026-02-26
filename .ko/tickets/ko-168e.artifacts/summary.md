# Implementation Summary: ko-168e

## What Was Done

Successfully removed the `project` field from the POST /ko endpoint payload in favor of using the `--project` flag in the argv array. This simplifies the serve API by eliminating special-case project resolution logic from the server.

### Changes Made

1. **cmd_serve.go (lines 358-361)**: Removed `Project` field from request struct
2. **cmd_serve.go (lines 385-421)**: Deleted 37 lines of project resolution logic including:
   - Registry lookup for #tag syntax
   - Absolute path handling
   - Setting cmd.Dir for project-scoped execution
3. **cmd_serve_test.go (line 53-56)**: Removed `Project` field from test handler request struct
4. **cmd_serve_test.go (line 162-167)**: Updated "empty project uses cwd" test to remove `"project":""` from payload
5. **cmd_serve_test.go (lines 542-747)**: Completely rewrote TestServeProjectScoped to test --project flag passthrough instead of server-side project resolution

### Key Architectural Decision

The server is now a simple passthrough that executes commands with their argv as-is. All project context handling happens client-side via the `--project` flag, which commands parse internally using their existing `resolveProjectTicketsDir()` logic.

### Test Strategy Change

The old TestServeProjectScoped created a full test environment with:
- Two temporary projects with .ko directories
- A test registry with #tag mappings
- Validation of server-side path resolution

The new test is simpler and validates that:
- The server correctly passes through argv containing --project flags
- Multiple flags (including --project) are preserved
- Commands without --project also pass through correctly

This aligns with the new architecture where the server doesn't interpret flags—it just executes them.

## Compliance

- ✅ All planned tasks completed
- ✅ No unexplained deviations from plan
- ✅ Tests pass (go test -run TestServe)
- ✅ No invariant violations detected
- ✅ Scope maintained (no unrelated changes)
- ✅ Simplification aligns with ticket goal

## Notes for Future Readers

- The removed project resolution logic (registry lookup, #tag parsing) still exists in individual commands via `resolveProjectTicketsDir()` (cmd_list.go:53-96)
- This change makes the serve endpoint stateless—it doesn't need access to the project registry
- If future callers need project-scoped execution, they must include `--project=<tag>` in the argv they send to the endpoint
