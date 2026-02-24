## Goal
Create a `.ko/.gitignore` file that excludes agent runtime files (agent.lock, agent.pid, agent.log) from version control, and have it automatically created by `ko agent init`.

## Context
The codebase already handles .gitignore creation for the `.ko/builds/` directory (see cmd_build_init.go:38-44). The pattern writes `*\n!.gitignore\n` to exclude all build artifacts while keeping the .gitignore file itself tracked.

Agent runtime files are created in `.ko/` by the agent daemon:
- `agent.lock` — Runtime flock for `ko agent loop` (cmd_loop.go:13-33)
- `agent.pid` — PID file for background agent (cmd_agent.go:49-51)
- `agent.log` — Log output (cmd_agent.go:54-56)

Currently, `.ko/agent.log` shows up as modified in git status, and these files shouldn't be tracked since they're transient runtime state.

The `cmdAgentInit` function in cmd_build_init.go creates the `.ko/` directory structure, prompts, and pipeline config. Tests for this command live in cmd_build_init_test.go.

## Approach
Add logic to `cmdAgentInit` to create `.ko/.gitignore` containing entries for the three runtime files. Follow the existing pattern used for `.ko/builds/.gitignore`. Update the test to verify the gitignore file is created.

## Tasks
1. [cmd_build_init.go:cmdAgentInit] — After creating the builds directory gitignore (line 44), add code to write `.ko/.gitignore` with entries for `agent.lock`, `agent.pid`, and `agent.log`.
   Verify: Run `ko agent init` in a test project, confirm `.ko/.gitignore` is created with correct contents.

2. [cmd_build_init_test.go:TestCmdBuildInit_createsFiles] — Add `.ko/.gitignore` to the list of expected files (around line 26-36).
   Verify: `go test -run TestCmdBuildInit_createsFiles` passes.

3. [cmd_build_init_test.go:TestCmdBuildInit_createsFiles] — After the existing file existence checks, add a verification that reads `.ko/.gitignore` and confirms it contains the three runtime file entries.
   Verify: `go test -run TestCmdBuildInit_createsFiles` passes with the new assertion.

## Open Questions
None — the implementation is straightforward following the existing `.ko/builds/.gitignore` pattern.
