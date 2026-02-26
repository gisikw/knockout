## Goal
Remove the `project` field from the POST /ko endpoint payload and rely solely on the --project flag in the argv array.

## Context
The serve endpoint at /ko currently accepts a JSON payload with two fields:
- `argv`: array of command arguments
- `project`: optional project context (string)

Lines 358-361 and 385-421 in cmd_serve.go parse this payload, resolve the project path (via registry lookup for #tags or treating as absolute path), then execute the command with cmd.Dir set to the resolved path.

The dependency ticket ko-6f77 introduced --project flag support to commands. With that change, commands like `ls`, `ready`, `agent`, etc. can now accept `--project=<tag>` as a flag in their argv. This means callers can specify: `{"argv": ["ls", "--project=foo"]}` instead of `{"argv": ["ls"], "project": "foo"}`.

The ticket author has confirmed (in Notes section) to "Remove payload" - meaning remove the special project field and let --project flow through argv like any other flag. This simplifies the API by removing a special case.

Current behavior:
- Request with `project` field → server resolves path → sets cmd.Dir → executes command
- Commands parse --project flag via resolveProjectTicketsDir() (cmd_list.go:53-96)

After this change:
- Request has no `project` field → server just executes argv as-is
- Commands handle --project flag internally (already implemented)

Tests affected:
- TestServeHandler (cmd_serve_test.go:162-167): has test "empty project uses cwd"
- TestServeProjectScoped (cmd_serve_test.go:542-747): extensively tests project resolution
- These tests validate the current project payload behavior and will need updating

## Approach
Remove the `project` field from the request struct and the project resolution logic from the /ko handler. Update tests to use --project in argv instead of the payload field. The server becomes a simpler passthrough that just executes commands without special project handling.

## Tasks
1. [cmd_serve.go:358-361] — Remove `Project` field from the request struct.
   Verify: Code compiles without errors.

2. [cmd_serve.go:385-421] — Delete the project resolution block (lines 385-421) that checks req.Project, resolves paths, and sets cmd.Dir.
   Verify: Handler now executes commands directly without path manipulation.

3. [cmd_serve_test.go:162-167] — Update test "empty project uses cwd" to remove the `"project":""` field from the payload.
   Verify: Test still passes (command runs in server's cwd).

4. [cmd_serve_test.go:542-747] — Replace TestServeProjectScoped with a simpler test that verifies --project flag passthrough. Instead of testing server-side project resolution, test that argv with --project flag executes correctly.
   Verify: New test validates commands receive --project in their argv.

5. [cmd_serve_test.go:53-56, 606-609] — Remove `Project` field from test handler's request struct in both TestServeHandler and the handler within TestServeProjectScoped test.
   Verify: All serve tests pass with `go test -run TestServe`.

## Open Questions
None. The approach is straightforward: remove server-side special handling of project context since commands now handle --project flags themselves. This is a simplification with no behavioral gaps.
