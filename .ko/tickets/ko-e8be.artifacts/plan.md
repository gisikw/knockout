## Goal
Create a simple, self-contained bash script (`jab.sh`) that demonstrates the core workflow pattern (plan → implement → test → validate) in a trivially portable way for users to customize.

## Context
The knockout system has a sophisticated workflow pipeline with decision/action nodes, multiple workflows, and YAML configuration. The ticket asks for the opposite: a single bash script that encapsulates a miniature version of this workflow pattern for educational/demo purposes.

Key patterns from the codebase:
- Agent harnesses like `agent-harnesses/claude.sh` invoke the Claude CLI via stdin: `echo "$KO_PROMPT" | claude $args`
- The harness uses simple flag-based invocation: `-p --output-format text`
- Prompts in `examples/` show a concise, instructional style (2-5 sentences, bullet rules)
- Workflows consist of typed nodes that execute prompts or bash commands
- The system is designed to be portable and text-based

The script should be:
- Single file, no external dependencies beyond bash and `claude` CLI
- Embeddable prompts (heredocs in the script body)
- Simple loop over input files
- Four stages: plan, implement, test (bash), validate
- Instructional/demo quality — meant to be modified, not production-grade
- Executable from the project root: `./jab.sh file1 file2 file3`

## Approach
Create `jab.sh` at the project root as a standalone bash script with embedded prompt templates. The script will accept file paths as arguments, loop through them, and run each through the four-stage workflow. Plan, implement, and validate stages will invoke Claude CLI directly via stdin (following the harness pattern). The test stage will run a simple bash validation command. The script will be heavily commented to explain the workflow pattern and encourage customization.

## Tasks
1. [jab.sh] — Create the main script structure: shebang, usage function, argument validation, and main loop over input files.
   Verify: Script is executable and prints usage when called without arguments.

2. [jab.sh] — Add four stage functions (stage_plan, stage_implement, stage_test, stage_validate) with skeleton implementations.
   Verify: Each function can be called and prints a stage header.

3. [jab.sh] — Implement stage_plan and stage_implement with embedded prompts using heredocs, invoking Claude CLI via stdin pattern: `echo "$prompt" | claude -p --output-format text`.
   Verify: Both stages successfully call Claude CLI and display output.

4. [jab.sh] — Implement stage_validate with an embedded prompt, following the same pattern as plan/implement.
   Verify: stage_validate calls Claude CLI and displays output.

5. [jab.sh] — Implement stage_test as a simple bash validation (check file exists, attempt a basic language-specific syntax check if detectable via extension).
   Verify: stage_test runs successfully and reports pass/fail clearly.

6. [jab.sh] — Wire up the full workflow in the main loop: for each file, run all four stages in sequence, print clear stage transitions, accumulate and display a final summary.
   Verify: `./jab.sh <somefile>` runs all stages and produces readable output with clear stage demarcation.

7. [jab.sh] — Add extensive comments throughout: explain what each section does, what to customize (prompts, test logic, stage order), and how the stages map to knockout's workflow concept.
   Verify: Script is self-documenting and reads as an educational example.

## Open Questions
None. The ticket explicitly requests simplicity and portability ("STUPID simple", "trivially portable", "go make it yours"), which guides all implementation decisions:

- **Test stage uses bash validation** (not an LLM call) — the ticket says "test could be bash"
- **Claude CLI only** — keeping it simple; users can modify the script to support other tools
- **Output shown inline** — no separate log files, just stdout during execution for immediate feedback
- **No error recovery** — if a stage fails, the script can stop or continue to the next file; implementing graceful continuation with summary at end
