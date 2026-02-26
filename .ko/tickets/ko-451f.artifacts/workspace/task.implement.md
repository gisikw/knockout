## Summary

All tasks completed successfully. Here's what was changed:

### Task 1: Removed `blocked [id]` command line
- **Location**: README.md:18
- **Change**: Removed the line documenting `ko blocked [id]` command from the Commands section

### Task 2: Removed `ko triage` command lines
- **Location**: README.md:26-29 (original)
- **Change**: Removed all four lines documenting the `ko triage` command and its flags (`--block`, `--questions`, `--answers`)

### Task 3: Removed JSON output lines
- **Location**: README.md:70-71 (original)
- **Change**: Removed references to `ko blocked --json` and `ko triage <id> --json` from the JSON output section

### Task 4: Updated Plan Questions section
- **Location**: README.md:400-418 (current)
- **Change**: Replaced all instances of `ko triage` with `ko update` in the four example commands:
  - Blocking with questions
  - Reading block reason and questions
  - Answering questions
  - Blocking with a reason

### Task 5: Added `ko update` documentation
- **Location**: README.md:25-31
- **Change**: Added comprehensive documentation for the `ko update` command showing:
  - Show block reason and questions (no flags)
  - `--block` flag for blocking with optional reason
  - `--questions` flag for adding questions (implicitly blocks)
  - `--answers` flag for answering questions (auto-unblock when done)
  - `--status` flag for setting ticket status
  - General metadata update flags (title, description, priority, tags, etc.)

### Task 6: Built and installed ko binary
- **Action**: Ran `just install`
- **Result**: Successfully built and installed ko binary to GOPATH/bin (version 4abd6a6)

All changes follow the plan exactly. No deviations were required. The README now accurately reflects the consolidated `ko update` command surface, with all references to the removed `ko triage` and `ko blocked` commands updated or removed.
