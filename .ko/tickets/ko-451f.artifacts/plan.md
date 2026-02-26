## Goal
Update README.md to reflect the ko update command consolidation and install the updated binary.

## Context
The ko update command has been implemented and consolidates multiple subsidiary commands:
- `ko triage` with `--block`, `--questions`, `--answers` flags has been removed (ko-0c94)
- `ko blocked` command has been removed
- `ko start`, `ko close`, `ko status` are now wrappers around `ko update` (ko-c7a8)

The README.md currently documents the old command surface with:
1. Lines 26-29: `ko triage` with subsidiary flags
2. Line 18: `ko blocked [id]` command
3. Lines 70-71: JSON output for `ko blocked` and `ko triage`
4. Lines 408-426: "Plan Questions" section with `ko triage` examples
5. Line 244: Pipeline config example using "triage" workflow node name

The `ko update` command now handles all these operations through a unified interface with flags like `--questions`, `--answers`, `--status`, `--block` (implicit through questions).

The Justfile has an `install` target that builds and installs ko to GOPATH/bin.

## Approach
Replace all references to the removed `ko triage` and `ko blocked` commands with the consolidated `ko update` command. Update the help text section, JSON output section, and Plan Questions examples. Leave the pipeline config example unchanged since workflow node names are user-defined (triage is a valid node name). Run `just install` to build and install the updated ko binary.

## Tasks
1. [README.md:18] — Remove the `blocked [id]` command line from the Commands section.
   Verify: Line no longer appears in README.

2. [README.md:26-29] — Remove the four `ko triage` command lines (show, --block, --questions, --answers).
   Verify: Lines no longer appear in README.

3. [README.md:70-71] — Remove the two JSON output lines for `ko blocked` and `ko triage`.
   Verify: Lines no longer appear in README.

4. [README.md:408-426] — Replace the "Plan Questions" section examples. Replace `ko triage` with `ko update` in all examples. The functionality is the same, just the command name changes.
   Verify: All examples use `ko update` instead of `ko triage`.

5. [README.md] — After status commands section (around line 24), add documentation for `ko update` command showing its flags and that it consolidates multiple operations. Use the same format as the current help output shows.
   Verify: New section matches format of existing command documentation.

6. Run `just install` to build and install the updated ko binary to GOPATH/bin.
   Verify: Command completes successfully, binary is installed.

## Open Questions
None — the ko update command is already implemented and the README changes are straightforward replacements of command names and examples.
