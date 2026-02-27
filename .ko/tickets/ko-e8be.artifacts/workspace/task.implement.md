Perfect! All tasks complete. Let me provide a summary of what was implemented.

---

## Implementation Summary

I successfully created `jab.sh` at the project root, implementing all 7 tasks from the plan:

### Task 1: Main script structure
- Created `jab.sh` with shebang (`#!/usr/bin/env bash` for NixOS compatibility)
- Added `usage()` function with clear examples and requirements
- Added argument validation (exits if no files provided)
- Set up main loop skeleton

### Task 2: Four stage functions
- Implemented `stage_plan()`, `stage_implement()`, `stage_test()`, and `stage_validate()` as separate functions
- Each prints a clear stage header using box-drawing characters

### Task 3: Plan and Implement stages
- Implemented both stages with embedded prompts using heredocs
- Followed the agent harness pattern: `echo "$prompt" | claude -p --output-format text`
- Prompts include file content inline (avoiding `--file` flag that requires authentication)
- Both stages successfully invoke Claude CLI and display output

### Task 4: Validate stage
- Implemented with embedded prompt following same pattern as plan/implement
- Asks Claude to review correctness, completeness, quality, and safety
- Successfully calls Claude CLI and displays output

### Task 5: Test stage
- Implemented as bash validation (no LLM call)
- Checks file existence first
- Runs language-specific syntax checks based on file extension:
  - Python: `python3 -m py_compile`
  - JavaScript: `node --check`
  - Bash: `bash -n`
  - Go: `go fmt`
- Reports pass/fail clearly with ✓, ✗, and ⊘ symbols

### Task 6: Full workflow integration
- Wired up all four stages in sequence using `&&` chaining
- Added clear visual demarcation:
  - Header banner with total file count
  - Per-file box with progress (FILE 1/2)
  - Stage transitions clearly marked
- Implemented summary section showing:
  - Total processed, successful, failed counts
  - List of failed files
  - Exit code 1 if any failures
- Graceful continuation: failures don't abort processing of remaining files

### Task 7: Extensive comments
- Added comprehensive header documentation explaining:
  - What the script does
  - Requirements
  - How it works (4 stages)
  - What a real workflow would do differently
- Added section headers for each major block (USAGE, STAGE 1-4, MAIN WORKFLOW)
- Each stage function has detailed comments explaining:
  - What type of stage it is (decision/action/bash)
  - How to customize it
  - What patterns it follows from knockout
- Main workflow has detailed comments mapping to knockout concepts
- Script reads as an educational example that encourages modification

### No deviations from plan
All tasks were completed as specified. The only minor adjustment was using `#!/usr/bin/env bash` instead of `#!/bin/bash` for NixOS compatibility, which is a standard portability practice and aligns with the plan's goal of being "trivially portable."

The script is functional, well-documented, and ready to be handed to users with the instruction: "here; it's not perfect, go make it yours."
