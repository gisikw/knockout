Perfect! The plan is complete. Let me provide a summary to the user:

I've investigated the codebase and written a plan to `$KO_ARTIFACT_DIR/plan.md`. Here's what I found:

## Key Findings

**Current Behavior:**
- `ko ls` filters out closed tickets by default (only one user explicitly uses `--status=closed` to see them)
- The SSE endpoints in `ko serve` return ALL tickets including closed ones

**What Needs to Change:**
Add an `--all` flag to `ko ls` that disables the default closed-ticket filter, providing parity with the SSE behavior.

## Implementation Plan

The plan includes 3 tasks:

1. **Add `--all` flag to cmdLs** - Modify the flag parsing and filtering logic in cmd_list.go
2. **Add spec** - Document the new behavior in specs/ticket_listing.feature  
3. **Add test** - Create a testscript test case in testdata/ticket_listing.txtar

This is a straightforward change that maintains backward compatibility (default behavior unchanged) while providing the requested functionality. The implementation follows project invariants: spec before code, pure decision functions, and maintains the file size constraint (cmd_list.go is currently 286 lines, well under the 500 line limit).
