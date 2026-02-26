Perfect! I've written a comprehensive plan for adding `--json` flag support to all commands that produce human-readable output. The plan:

## Summary

**Goal**: Add `--json` flag to all output commands for machine consumption.

**Context**: Found that `show`, `ls`, `ready`, and `blocked` already have JSON support. Four commands need it added:
- `agent status` - status information
- `triage` - block reason and questions  
- `dep tree` - dependency tree structure
- `project ls` - registered projects list

**Approach**: Follow the existing pattern of adding a `--json` flag using Go's flag package, defining appropriate struct types, and outputting JSON when the flag is set.

**Tasks**: 9 ordered tasks covering:
1. Implementation changes to each command file
2. Corresponding test additions/updates
3. Documentation updates in README.md

Each task includes the specific file and function to modify, what to change, and how to verify the change works.

**Open Questions**: None - the pattern is well-established in the codebase and requirements are clear.
