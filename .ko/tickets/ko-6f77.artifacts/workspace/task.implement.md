## Summary

I have successfully implemented all tasks from the plan. Here's what was changed:

### Task 1: Modified resolveProjectTicketsDir
- Changed from parsing hashtags to parsing `--project` flag
- Implemented manual flag parsing to avoid consuming other command flags
- Returns tickets directory and remaining args with the flag removed
- Falls back to local tickets directory if no `--project` flag is provided

**Files modified:** `cmd_list.go:53-89`

### Tasks 2-4: Updated command flag handling
- Added `--project` to reorderArgs maps in `cmdLs`, `cmdReady`, and `cmdAgentLoop`
- Ensures `--project` is parsed early before calling resolveProjectTicketsDir

**Files modified:** `cmd_list.go:92-96, 181-186`, `cmd_loop.go:83-87`

### Tasks 5-8: Agent commands
- Verified that `cmdAgentStart`, `cmdAgentStop`, `cmdAgentStatus`, and `cmdAgentInit` already pass args through resolveProjectTicketsDir correctly
- No changes needed as they already work with the new implementation

**Files verified:** `cmd_agent.go`, `cmd_build_init.go`

### Task 9: Removed RouteTicket from cmdCreate
- Removed call to `RouteTicket()` function
- Added `--project` flag to cmdCreate's flag set
- Replaced hashtag parsing with explicit `--project` flag logic
- Simplified ticket creation to no longer create audit tickets for routed tickets
- Title is now used as-is (hashtags are preserved literally in the title)

**Files modified:** `cmd_create.go:20-191`

### Task 10: Marked RouteTicket as deprecated
- Added deprecation notice to RouteTicket function documentation
- Kept function for backward compatibility

**Files modified:** `registry.go:301-306`

### Task 11: Wrote tests for --project flag
- Created `cmd_list_test.go` with comprehensive tests for resolveProjectTicketsDir
- Tests cover: no flag (local fallback), with flag (project routing), unknown project (error), and missing tickets dir (error)
- All new tests pass

**Files created:** `cmd_list_test.go`

## Discrepancy Found

The implementation removed hashtag-based routing as requested, which causes 3 existing tests in `testdata/project_registry/` to fail:
- `add_routed.txtar` - Tests hashtag routing (e.g., `ko add 'Fix #target'`)
- `add_captured.txtar` - Tests unrecognized hashtag routing to default project
- `add_multitag.txtar` - Tests multiple hashtags

These tests validate the old behavior that was intentionally removed. The plan does not mention updating or removing these tests. This is expected since the feature they test no longer exists. These tests need to be either:
1. Updated to use the new `--project` flag syntax, or
2. Removed if the old behavior is no longer supported

All other tests pass successfully.
