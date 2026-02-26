## Goal
Create a test scenario that validates the plan questions workflow in the knockout pipeline.

## Context
The knockout system supports structured plan questions via the `needs_input` disposition (disposition.go:11-28). When the actionable gate (actionable.md) encounters open questions in a plan, it converts them into structured PlanQuestion objects with IDs, question text, context, and options. These questions are stored in the ticket's YAML frontmatter (ticket.go:35,46-58) and the ticket is blocked until they're answered via `ko update --answers`.

The build pipeline processes `needs_input` dispositions by copying `disp.PlanQuestions` to `t.PlanQuestions` and setting status to blocked (build.go:335-338).

This ticket is explicitly testing that behavior by requesting a plan with open questions.

## Approach
Since this is a meta-ticket testing the plan questions feature itself, the implementation would involve creating a contrived scenario that legitimately requires human input. However, the actual requirement is simpler: demonstrate that a plan can properly surface open questions that the actionable gate will catch.

The most authentic approach is to propose a feature that genuinely has multiple valid implementation paths with meaningful tradeoffs that cannot be resolved from the codebase alone.

## Tasks
1. No implementation tasks — this ticket exists to test the planning phase behavior.
   Verify: The actionable gate correctly converts open questions to structured PlanQuestion objects.

## Open Questions

**Q1: What type of test scenario should be implemented?**

The system supports multiple workflows (task, research, bug) and various disposition types. Without knowing the specific behavior being tested, there are several valid approaches:

- **Option A: Add a new workflow node type** — Would test the full pipeline including disposition parsing and workflow routing. Requires modifying pipeline config schema and node execution logic.

- **Option B: Add a CLI command for inspecting plan questions** — Would test the ticket data model and serialization. Simpler scope, focused on the data structures rather than pipeline orchestration.

- **Option C: Create end-to-end test suite** — Would test the entire flow from plan generation through question answering to unblocking. Most comprehensive but largest scope.

**Q2: Should the test be interactive or automated?**

The ticket says "I'm trying to test that behavior" but doesn't specify whether this is:

- **Manual verification** — Run the pipeline on this ticket, observe that questions are generated, then manually answer them to verify the unblocking flow.

- **Automated test** — Add a Go test case that simulates the plan → actionable → needs_input → answer flow programmatically.

**Q3: What should happen after questions are answered?**

Once the human answers the questions via `ko update --answers`, the ticket unblocks. But then what?

- **Close immediately** — Treat this as a test-only ticket that's done once the workflow is validated.

- **Implement a small feature** — Follow through with whatever implementation path is chosen based on the answers, proving the full cycle works.

- **Document findings** — Write up the test results and close as a research ticket.
