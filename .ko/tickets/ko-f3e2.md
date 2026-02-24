---
id: ko-f3e2
status: closed
deps: [ko-8a7c]
links: []
created: 2026-02-24T03:04:21Z
type: task
priority: 2
---
# Plan node: write plan artifact, block with questions if needed

Write the `plan.md` prompt for the task workflow's plan node.

The plan node explores the codebase and produces a structured plan artifact
at `$KO_ARTIFACT_DIR/plan.md`. If the plan has open questions that require
human input, it blocks the ticket with a structured reason.

Plan output format:
- Goal (one sentence)
- Context (relevant files found, patterns, conventions)
- Approach (2-3 sentences)
- Tasks (ordered, each with file refs and verification step)
- Risks / open questions

Key behaviors:
- Check for existing plan artifact from a prior attempt before re-crawling
- If open questions exist that the agent can't resolve: `fail` with
  specific questions listed (these become the "what do I need from you"
  annotations for the overnight queue use case)
- Write triage context to artifact dir so implement doesn't re-crawl
- No code in the plan — natural language with file references

Depends on: ko-8a7c (artifact directory)
