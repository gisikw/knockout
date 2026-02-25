## Goal
Add optional project field to POST /ko endpoint to execute commands in a specific project directory.

## Context
The POST /ko endpoint (cmd_serve.go:354-406) currently executes ko subcommands in the server's working directory using `exec.Command(os.Args[0], req.Argv...)`. The subscribe endpoint already implements project resolution (lines 220-254): it accepts a project parameter either as a query param or path segment, resolves #tags via registry lookup, and uses absolute paths directly.

The codebase has an established pattern for setting working directory on exec.Cmd instances (build.go:674 and cmd_agent.go:147 both use `cmd.Dir = projectRoot`).

The registry module (registry.go) provides:
- `RegistryPath()` to locate the registry file
- `LoadRegistry(path)` to load project mappings
- Registry lookups via `reg.Projects[tag]` map

The spec (http_server.feature) describes security constraints and whitelist enforcement but does not yet document project-scoped execution.

## Approach
Add a "project" field to the POST /ko request body struct (alongside "argv"). When present, resolve it using the same logic as the subscribe endpoint: #tag syntax gets registry lookup, otherwise treat as absolute path. Set cmd.Dir to the resolved project path before exec. When absent, preserve current behavior (inherit server's cwd).

## Tasks
1. [cmd_serve.go:361-363] — Add optional `Project string` field to the request struct.
   Verify: code compiles.

2. [cmd_serve.go:387-389] — After whitelist check, add project resolution logic: if req.Project is non-empty and starts with "#", load registry and resolve tag to path; if non-empty without "#", use as absolute path; set cmd.Dir to resolved path.
   Verify: logic mirrors subscribe endpoint's resolution pattern (lines 231-254).

3. [cmd_serve.go:388] — Set `cmd.Dir = resolvedProjectPath` before calling CombinedOutput when project is specified.
   Verify: follows established pattern from build.go:674 and cmd_agent.go:147.

4. [cmd_serve_test.go:56-57] — Add Project field to test request struct definition.
   Verify: test compilation succeeds.

5. [cmd_serve_test.go:163] — Add test cases for project parameter: valid #tag, absolute path, unrecognized tag (should error), empty string (should use cwd).
   Verify: `go test -run TestServeHandler` passes with new cases.

6. [specs/http_server.feature:170-178] — Add new Rule section documenting project-scoped execution with scenarios: request with #tag, request with absolute path, request with invalid tag returns error, request without project uses cwd.
   Verify: spec documents the new behavior clearly.

## Open Questions
None — the subscribe endpoint's resolution pattern is the clear reference implementation to follow, and cmd.Dir setting is already established in the codebase.
