## Goal
Create a simple, self-contained bash script (`jab.sh`) that demonstrates the core workflow pattern (plan → implement → test → validate) in a trivially portable way for users to customize.

## Context
The knockout system has a sophisticated workflow pipeline with decision/action nodes, multiple workflows, and YAML configuration. The ticket asks for the opposite: a single bash script that encapsulates a miniature version of this workflow pattern for educational/demo purposes.

Key patterns from the codebase:
- Workflows consist of typed nodes (decision vs action) that execute prompts or bash commands
- Agent harnesses (like `agent-harnesses/claude.sh`) show how to invoke LLMs via environment variables
- Prompts live in `.ko/prompts/` as markdown files, but for portability they need to be embedded in the script
- The pipeline uses `$KO_PROMPT` to pass prompt text to harnesses

The script should be:
- Single file, no external dependencies beyond bash and `claude` CLI
- Embeddable prompts (heredocs in the script body)
- Simple loop over input files
- Four stages: plan, implement, test (bash), validate
- Instructional/demo quality — meant to be modified, not production-grade

## Approach
Create `jab.sh` as a standalone bash script with embedded prompt templates. The script will accept file paths as arguments, loop through them, and run each through the four-stage workflow. Each stage will either invoke the Claude CLI with an embedded prompt (plan, implement, validate) or run a simple bash test command. Output from each stage will be visible for debugging. The script will be heavily commented to encourage modification.

## Tasks
1. [jab.sh] — Create the main script structure with argument parsing and file loop.
   Verify: Script can be invoked with `./jab.sh file1 file2` and prints usage when called without arguments.

2. [jab.sh] — Add four stage functions (stage_plan, stage_implement, stage_test, stage_validate) with embedded prompts using heredocs.
   Verify: Each function can be called independently and produces expected output format.

3. [jab.sh] — Implement Claude CLI invocation pattern based on `agent-harnesses/claude.sh`, but simplified (direct invocation, no complex harness features).
   Verify: `stage_plan` and `stage_implement` successfully call Claude CLI and capture output.

4. [jab.sh] — Add the test stage as a simple bash validation (e.g., file exists, syntax check for common languages).
   Verify: `stage_test` runs a basic test command and returns success/failure.

5. [jab.sh] — Wire up the full workflow loop: for each file, run all four stages in sequence, show stage transitions, handle errors gracefully.
   Verify: `./jab.sh testfile.py` runs all stages and produces readable output.

6. [jab.sh] — Add extensive comments explaining each section, what to customize, and how the stages map to knockout's workflow concept.
   Verify: Script is self-documenting and contains clear customization hints.

## Open Questions
**Assumption: Test stage should be a placeholder bash command** (like checking file exists, or running a language-specific linter if the file type is detectable). The ticket says "test could be bash" which I interpret as "test is just a bash command, not an LLM call." If it should also be an LLM prompt, that's a quick adjustment.

**Assumption: Claude CLI is the target LLM tool.** The codebase uses `claude` in its harness. If the script should support other tools (like `cursor`), we'd need to add detection/configuration logic, which conflicts with "STUPID simple." Sticking with claude-only for simplicity.

**Assumption: Each stage overwrites previous output** (or appends to a log). The ticket doesn't specify if stage outputs should be preserved separately or accumulated. Will default to showing output inline during execution, not persisting to files, since that's simpler and more demo-friendly.
