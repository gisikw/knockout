Perfect! Now let me provide a summary of all the changes made.

## Summary

I've successfully implemented `--json` flag support for all output commands that needed it. Here's what was changed, organized by task:

### Task 1-2: Agent Status Command
**Files Modified:** `cmd_agent.go`, `cmd_agent_test.go` (created)
- Added `agentStatusJSON` struct with fields for `provisioned`, `running`, `pid`, and `last_log`
- Modified `cmdAgentStatus` to accept `--json` flag
- Created comprehensive tests covering provisioned/not provisioned, running/not running, and stale pid file scenarios

### Task 3-4: Triage Command
**Files Modified:** `cmd_triage.go`, `cmd_triage_test.go`
- Added `triageStateJSON` struct with fields for `block_reason` and `questions`
- Modified `cmdTriage` and `showTriageState` to support `--json` flag
- Added tests verifying JSON output for blocked tickets with/without reasons and questions

### Task 5-6: Dep Tree Command
**Files Modified:** `cmd_dep.go`, `cmd_dep_test.go` (created)
- Added `depTreeJSON` struct with recursive `deps` field
- Implemented `buildDepTree` pure function to construct tree structure (handles cycles)
- Modified `cmdDepTree` to support `--json` flag with proper flag reordering
- Created tests covering single node, linear chain, branching tree, and cycle detection

### Task 7-8: Project Ls Command
**Files Modified:** `cmd_project.go`, `cmd_project_test.go`
- Added `projectJSON` struct with fields for `tag`, `path`, and `is_default`
- Modified `cmdProjectLs` to support `--json` flag
- Added tests verifying JSON output for empty registry, multiple projects, and default markers

### Task 9: Documentation
**Files Modified:** `README.md`
- Added new "JSON output" section documenting all commands that support `--json`
- Listed all 8 commands: `show`, `ls`, `ready`, `blocked`, `triage`, `agent status`, `dep tree`, and `project ls`

### Verification
- All existing tests continue to pass
- All new tests pass
- Code builds successfully
- No deviations from the plan

The implementation follows existing patterns in the codebase, uses table-driven tests, and maintains the 500-line file size limit (all modified files remain under this threshold).
