Feature: Build Pipeline
  As an automated system
  I want to run build pipelines against tickets
  So that the ready queue is systematically burned down

  # Pipeline configuration

  Scenario: Pipeline is loaded from .ko/pipeline.yml
    Given a project with a pipeline config at ".ko/pipeline.yml"
    And a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then the pipeline should load from ".ko/pipeline.yml"

  Scenario: Missing pipeline config is a fatal error
    Given a project with no pipeline config
    And a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "no pipeline config found"

  Scenario: Pipeline config declares named workflows with typed nodes
    Given a pipeline with workflows: main, hotfix
    Then the "main" workflow has nodes: triage (decision), implement (action)
    And each node has a name and a type (decision or action)
    And each prompt node has a prompt file reference
    And each run node has a shell command
    And prompt and run are mutually exclusive per node

  Scenario: Pipeline must have a "main" workflow
    Given a pipeline with no "main" workflow
    When the pipeline is parsed
    Then parsing should fail with "must have a 'main' workflow"

  Scenario: Pipeline config specifies a default command for prompt nodes
    Given a pipeline with "command: my-llm-tool"
    Then prompt nodes invoke "my-llm-tool" instead of the default
    # Default command is "claude" but configurable for testing and future tools

  # Node execution

  Scenario: Nodes execute sequentially within a workflow
    Given a "main" workflow with nodes: triage, implement
    And a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then triage runs before implement

  Scenario: Prompt node invokes the configured command with ticket context
    Given a ticket "ko-a001" with status "open"
    And a "main" workflow with a prompt node "triage" using "triage.md"
    When the triage node runs
    Then the command receives the ticket content
    And the command receives the prompt file content
    And the command receives the discretion level

  Scenario: Run node executes a shell command
    Given a ticket "ko-a001" with status "open"
    And a "main" workflow with a run node "verify" using "just test"
    When the verify node runs
    Then "just test" is executed as a shell command
    And the exit code determines success or failure

  Scenario: Action node output is not parsed
    Given an action node that produces arbitrary output
    When the node completes successfully
    Then the output is saved to the workspace but not parsed for dispositions

  # Decision node dispositions

  Scenario: Decision node returns continue disposition
    Given a decision node that outputs '{"disposition": "continue"}'
    When the node runs
    Then execution advances to the next node in the workflow

  Scenario: Decision node returns fail disposition
    Given a ticket "ko-a001" with status "open"
    And a decision node that outputs '{"disposition": "fail", "reason": "Cannot implement"}'
    When the node runs
    Then ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "FAIL"
    And no retries are attempted (valid dispositions are never retried)

  Scenario: Decision node returns blocked disposition
    Given a ticket "ko-a001" with status "open"
    And a ticket "ko-b002" with status "open"
    And a decision node that outputs '{"disposition": "blocked", "block_on": "ko-b002", "reason": "Needs auth first"}'
    When the node runs
    Then ticket "ko-a001" should have "ko-b002" in deps
    And ticket "ko-a001" should have a note containing "BLOCKED"

  Scenario: Decision node returns decompose disposition
    Given a ticket "ko-a001" with status "open"
    And a decision node that outputs '{"disposition": "decompose", "subtasks": ["Sub one", "Sub two", "Sub three"]}'
    When the node runs
    Then 3 child tickets should exist with IDs starting with "ko-a001."
    And ticket "ko-a001" should depend on all 3 children
    And ticket "ko-a001" should not appear in ready

  Scenario: Decision node returns route disposition
    Given a "main" workflow with a decision node "triage" that declares routes: [feature]
    And a "feature" workflow with an action node "implement"
    And the triage node outputs '{"disposition": "route", "workflow": "feature"}'
    When the node runs
    Then execution jumps to the "feature" workflow
    And the "feature" workflow runs to completion

  Scenario: Route to undeclared workflow is a build failure
    Given a decision node "triage" that declares routes: [feature]
    And the triage node outputs '{"disposition": "route", "workflow": "hotfix"}'
    When the node runs
    Then the build fails
    And the ticket is blocked with a note about the undeclared route

  Scenario: Disposition is extracted from the last fenced JSON block
    Given a decision node that outputs:
      """
      Here is my analysis...

      ```json
      {"disposition": "fail", "reason": "first block ignored"}
      ```

      Actually wait, let me reconsider.

      ```json
      {"disposition": "continue"}
      ```
      """
    When the disposition is extracted
    Then the disposition is "continue" (from the last block)

  # Outcomes — every outcome removes the ticket from ready

  Scenario: SUCCEED closes the ticket
    Given a ticket "ko-a001" with status "open"
    And a pipeline where all nodes succeed
    When I run "ko build ko-a001"
    Then the command should succeed
    And ticket "ko-a001" should have status "closed"
    And ticket "ko-a001" should have a note containing "SUCCEED"

  Scenario: FAIL marks the ticket as blocked
    Given a ticket "ko-a001" with status "open"
    And a pipeline where a node fails after retries
    When I run "ko build ko-a001"
    Then the command should fail
    And ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "FAIL"

  # Eligibility

  Scenario: Only open tickets can be built
    Given a ticket "ko-a001" with status "in_progress"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "not eligible"

  Scenario: Closed tickets cannot be built
    Given a ticket "ko-a001" with status "closed"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "already closed"

  Scenario: Blocked tickets cannot be built
    Given a ticket "ko-a001" with status "blocked"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "blocked"

  Scenario: Tickets with unresolved deps cannot be built
    Given a ticket "ko-a001" with status "open" depending on "ko-b002"
    And a ticket "ko-b002" with status "open"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "unresolved dependencies"

  # Ticket status during build

  Scenario: Ticket is marked in_progress during build
    Given a ticket "ko-a001" with status "open"
    When "ko build ko-a001" starts executing nodes
    Then ticket "ko-a001" should have status "in_progress"

  # Visit limits

  Scenario: Node visit count is bounded
    Given a decision node "triage" with max_visits: 2
    And a workflow that routes back to the same workflow
    When triage is entered a third time
    Then the build fails with "exceeded max_visits"
    And the ticket is blocked

  # Retry logic

  Scenario: Failed node is retried up to max_retries
    Given a pipeline with max_retries: 2
    And a node that fails on first attempt but succeeds on second
    When the node runs
    Then the node should be attempted 2 times total
    And the pipeline should continue

  Scenario: Node failure after all retries is a build failure
    Given a pipeline with max_retries: 2
    And a node that fails on all attempts
    When the node runs
    Then the node should be attempted 3 times total
    And the outcome should be FAIL

  Scenario: Invalid disposition JSON is retry-eligible
    Given a decision node that produces output without valid fenced JSON
    And a pipeline with max_retries: 1
    When the node runs
    Then the node is retried
    # Valid dispositions (even "fail") are never retried

  Scenario: Valid disposition signals are not retried
    Given a decision node that outputs '{"disposition": "fail", "reason": "reason"}'
    When the node runs
    Then no retries are attempted
    And the outcome is FAIL immediately

  # Workspace

  Scenario: Build creates a workspace directory
    Given a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then a workspace directory exists under ".ko/builds/"
    And node outputs are tee'd as "<workflow>.<node>.md"
    And $KO_TICKET_WORKSPACE is set for all nodes and hooks

  # Lifecycle hooks

  Scenario: on_succeed runs after all workflows pass, before ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_succeed hooks
    When all workflows pass
    Then on_succeed hooks run
    And then ticket "ko-a001" is closed

  Scenario: on_close runs after ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_close hooks
    When "ko build ko-a001" succeeds
    Then ticket "ko-a001" is closed first
    And then on_close hooks run

  Scenario: Hook commands receive TICKET_ID, CHANGED_FILES, and KO_TICKET_WORKSPACE
    Given a pipeline with on_succeed: "echo ${TICKET_ID} ${CHANGED_FILES}"
    When on_succeed runs for ticket "ko-a001"
    Then TICKET_ID is set to "ko-a001"
    And CHANGED_FILES contains files modified during the build
    And KO_TICKET_WORKSPACE points to the build workspace

  Scenario: on_succeed failure prevents ticket close
    Given a ticket "ko-a001" with status "open"
    And a pipeline with an on_succeed hook that fails
    When "ko build ko-a001" runs
    Then ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "on_succeed failed"

  # CHANGED_FILES tracking

  Scenario: CHANGED_FILES captures files modified during build
    Given a ticket "ko-a001" with status "open"
    And a pipeline where the implement node modifies "foo.go" and creates "bar.go"
    When the build completes
    Then CHANGED_FILES should contain "foo.go" and "bar.go"

  # Decomposition depth guard

  Scenario: Decomposition is denied at max depth
    Given a ticket "ko-a001.b002" at depth 1
    And a pipeline with max_depth: 1
    And a decision node that outputs '{"disposition": "decompose", "subtasks": ["Subtask"]}'
    When the node runs
    Then ticket "ko-a001.b002" should have status "blocked"
    And ticket "ko-a001.b002" should have a note containing "max decomposition depth"
    And no child tickets should have been created

  Scenario: Decomposition is allowed below max depth
    Given a ticket "ko-a001" at depth 0
    And a pipeline with max_depth: 2
    And a decision node that outputs '{"disposition": "decompose", "subtasks": ["Sub one", "Sub two"]}'
    When the node runs
    Then 2 child tickets should exist with IDs starting with "ko-a001."

  # External ask limit

  Scenario: Build can create at most one external ask per run
    Given a ticket "ko-a001" with status "open"
    And a pipeline where a node requests 2 external asks
    When "ko build ko-a001" runs
    Then exactly 1 external ask should have been created
    And ticket "ko-a001" should have a note about throttled external asks

  # Build artifacts

  Scenario: Build creates an artifact directory
    Given a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then a build directory should exist under ".ko/builds/"
    And it should contain the ticket snapshot
    And it should contain output from each node

  # Loop safety invariant

  Scenario: No outcome leaves ticket unchanged on ready queue
    Given a ticket "ko-a001" with status "open"
    When "ko build ko-a001" completes with any outcome
    Then ticket "ko-a001" should not appear in "ko ready" output

  # Discretion levels

  Scenario: Discretion level is passed to prompt nodes
    Given a pipeline with discretion: high
    And a prompt node
    When the node runs
    Then the prompt context includes the discretion level "high"
    And the discretion guidance text for "high" is included

  # Model resolution

  Scenario: Model resolves with node > workflow > pipeline precedence
    Given a pipeline with model: "pipeline-model"
    And a workflow with model: "workflow-model"
    And a node with model: "node-model"
    Then the node uses "node-model"
    # If node has no model, workflow model is used
    # If workflow has no model, pipeline model is used

  # build-init

  Scenario: build-init scaffolds pipeline config
    Given a project with .tickets/ but no .ko/
    When I run "ko build-init"
    Then .ko/pipeline.yml should exist
    And .ko/prompts/triage.md should exist
    And .ko/prompts/implement.md should exist
    And .ko/prompts/review.md should exist
    And the generated pipeline.yml should be valid

  Scenario: build-init refuses to overwrite existing config
    Given a project with an existing .ko/pipeline.yml
    When I run "ko build-init"
    Then the command should fail
    And the error should contain "already exists"
    And the existing .ko/pipeline.yml should be unchanged
