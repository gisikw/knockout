Perfect! I've created a plan for implementing `jab.sh`. The plan outlines:

**Goal:** A simple, portable bash script demonstrating the workflow pattern (plan → implement → test → validate) that users can easily customize.

**Key decisions:**
- Single-file bash script with embedded prompts (no external config files)
- Uses heredocs for prompt templates
- Loops through files passed as arguments
- Four stages: three LLM calls (plan, implement, validate) + one bash test
- Heavily commented for educational value

**Open questions captured:**
1. Test stage is assumed to be a simple bash command (not LLM) — based on "test could be bash" in the ticket
2. Targets Claude CLI only (not multiple tools) to keep it simple
3. Output shown inline rather than persisted to separate files

The plan follows the required format with clear tasks, verification steps, and proportional scope for what is essentially a demo/educational script.
