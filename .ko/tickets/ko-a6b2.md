---
id: ko-a6b2
status: blocked
deps: []
created: 2026-02-26T04:49:28Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "What type of test scenario should be implemented?"
    context: "The knockout system supports multiple workflows and disposition types. This choice determines the scope and what aspects of the pipeline are validated."
    options:
      - label: "New workflow node type"
        value: new_workflow_node
        description: "Test the full pipeline including disposition parsing and workflow routing, but requires modifying pipeline config and node execution logic"
      - label: "CLI command for inspecting plan questions"
        value: cli_command
        description: "Test the ticket data model and serialization with simpler scope, focused on data structures rather than pipeline orchestration"
      - label: "End-to-end test suite"
        value: test_suite
        description: "Test the entire flow from plan generation through question answering to unblocking, most comprehensive but largest scope"
  - id: q2
    question: "Should the test be interactive or automated?"
    context: "The ticket requests testing the plan questions workflow behavior. This affects how the test validates the system."
    options:
      - label: "Manual verification"
        value: manual
        description: "Run the pipeline on this ticket, observe questions are generated, then manually answer them to verify unblocking"
      - label: "Automated test"
        value: automated
        description: "Add a Go test case that simulates the full plan → actionable → needs_input → answer flow programmatically"
  - id: q3
    question: "What should happen after questions are answered?"
    context: "Once the human answers questions via `ko update --answers`, the ticket unblocks. This determines the full lifecycle."
    options:
      - label: "Close immediately"
        value: close_on_answer
        description: "Treat this as a test-only ticket that's done once the workflow is validated"
      - label: "Implement based on answers"
        value: implement_feature
        description: "Follow through with whatever implementation path is chosen, proving the full cycle works"
      - label: "Document findings"
        value: document_findings
        description: "Write up the test results and close as a research ticket"
---
# I'd really like this ticket to end up with questions in its plan, as I'm trying to test that behavior
