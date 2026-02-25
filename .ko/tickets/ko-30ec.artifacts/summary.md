# Implementation Summary: ko-30ec

## What Was Done

Added optional `project` field to the POST /ko endpoint's JSON body to enable project-scoped command execution. The implementation mirrors the existing project resolution pattern from the subscribe endpoint.

### Changes Made

1. **cmd_serve.go:361-363** — Added `Project string` field to request struct alongside `Argv`

2. **cmd_serve.go:387-433** — Implemented project resolution logic:
   - If `project` starts with `#`, performs registry lookup to resolve tag to absolute path
   - If `project` is non-empty without `#`, treats as absolute path
   - If `project` is empty or absent, uses server's cwd (preserves existing behavior)
   - Returns 404 with JSON error for unrecognized tags
   - Returns 500 with JSON error for registry errors
   - Sets `cmd.Dir` to resolved path before execution

3. **cmd_serve_test.go:56-57** — Updated test request struct to include `Project` field

4. **cmd_serve_test.go:163-169** — Added test case for empty project string

5. **cmd_serve_test.go:546-753** — Added comprehensive `TestServeProjectScoped` with 5 test cases:
   - #tag resolution to registered project
   - Absolute path usage
   - Invalid tag error handling (404 response)
   - Empty project string uses cwd
   - Missing project field uses cwd

6. **specs/http_server.feature:180-236** — Added "Project-scoped execution" rule with 5 scenarios documenting all behaviors

## Key Decisions

- **Error handling:** Invalid tags return 404 (not found) with JSON error body, registry errors return 500 with JSON error body. This matches HTTP semantics and provides structured error responses for API consumers.

- **Backwards compatibility:** When `project` field is absent or empty, behavior is unchanged (uses server's cwd). This ensures existing clients continue to work.

- **Test structure:** Created a dedicated `TestServeProjectScoped` function with its own test handler rather than expanding `TestServeHandler`, keeping test concerns separated and making the project-resolution test setup explicit.

- **Spec organization:** Added as a new "Rule" section in the existing `http_server.feature` rather than creating a separate spec file, keeping all POST /ko behavior documented together.

## Compliance Notes

- **Specs and tests:** Added both feature spec scenarios and corresponding Go tests per INVARIANTS.md requirement
- **Error handling:** Structured JSON errors for API failures, clear HTTP status codes (404 for not found, 500 for registry errors)
- **Code organization:** Implementation mirrors the subscribe endpoint's resolution pattern (lines 231-254), maintaining consistency
- **Security:** Project resolution validates against registry for #tags, preventing arbitrary path traversal beyond what absolute paths already allow

## Test Results

All tests pass:
- `TestServeHandler` (10 cases including new empty project case)
- `TestServeProjectScoped` (5 cases covering all project resolution scenarios)
- Full test suite: `go test ./...` passes

## Notes for Future Readers

The subscribe endpoint (cmd_serve.go:220-254) was the reference implementation for project resolution. Any future changes to project resolution logic should be applied consistently to both endpoints.

The implementation handles registry errors gracefully but does not validate that resolved paths actually exist or are valid knockout projects — that's left to the ko subcommand execution, which will fail appropriately if the path is invalid.
