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

  Scenario: Pipeline config specifies stages
    Given a pipeline with stages: triage, implement, verify, review
    Then each stage has a name
    And each prompt stage has a prompt file reference
    And each run stage has a shell command
    And prompt and run are mutually exclusive per stage

  Scenario: Pipeline config specifies a default command for prompt stages
    Given a pipeline with "command: my-llm-tool"
    Then prompt stages invoke "my-llm-tool" instead of the default
    # Default command is "claude" but configurable for testing and future tools

  # Stage execution

  Scenario: Stages execute sequentially
    Given a pipeline with stages: triage, implement
    And a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then triage runs before implement
    And the output of triage is available to implement

  Scenario: Prompt stage invokes the configured command with ticket context
    Given a ticket "ko-a001" with status "open"
    And a pipeline with a prompt stage "triage" using "triage.md"
    When the triage stage runs
    Then the command receives the ticket content
    And the command receives the prompt file content
    And the command receives the discretion level

  Scenario: Run stage executes a shell command
    Given a ticket "ko-a001" with status "open"
    And a pipeline with a run stage "verify" using "just test"
    When the verify stage runs
    Then "just test" is executed as a shell command
    And the exit code determines success or failure

  Scenario: Stage output chains forward
    Given a pipeline with stages: triage, implement
    When triage produces output "Files to modify: foo.go"
    Then implement receives "Files to modify: foo.go" as previous stage output

  # Outcomes — every outcome removes the ticket from ready

  Scenario: SUCCEED closes the ticket
    Given a ticket "ko-a001" with status "open"
    And a pipeline where all stages succeed
    When I run "ko build ko-a001"
    Then the command should succeed
    And ticket "ko-a001" should have status "closed"
    And ticket "ko-a001" should have a note containing "SUCCEED"

  Scenario: FAIL marks the ticket as blocked (HITL)
    Given a ticket "ko-a001" with status "open"
    And a pipeline where a stage fails after retries
    When I run "ko build ko-a001"
    Then the command should fail
    And ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "FAIL"
    And ticket "ko-a001" should have a note containing the failure reason

  Scenario: Prompt stage signals FAIL explicitly
    Given a ticket "ko-a001" with status "open"
    And a prompt stage that outputs "FAIL\nCannot implement: missing API spec"
    When the stage runs
    Then the outcome is FAIL with reason "Cannot implement: missing API spec"
    And no retries are attempted

  Scenario: BLOCKED wires a dependency
    Given a ticket "ko-a001" with status "open"
    And a ticket "ko-b002" with status "open"
    And a prompt stage that outputs "BLOCKED ko-b002\nNeeds auth refactor first"
    When the stage runs
    Then ticket "ko-a001" should have "ko-b002" in deps
    And ticket "ko-a001" should have a note containing "BLOCKED"

  Scenario: DECOMPOSE creates children and blocks parent
    Given a ticket "ko-a001" with status "open"
    And a prompt stage that outputs "DECOMPOSE\n- Subtask one\n- Subtask two\n- Subtask three"
    When the stage runs
    Then 3 child tickets should exist with IDs starting with "ko-a001."
    And ticket "ko-a001" should depend on all 3 children
    And ticket "ko-a001" should not appear in ready

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
    When "ko build ko-a001" starts executing stages
    Then ticket "ko-a001" should have status "in_progress"

  # Retry logic

  Scenario: Failed stage is retried up to max_retries
    Given a pipeline with max_retries: 2
    And a stage that fails on first attempt but succeeds on second
    When the stage runs
    Then the stage should be attempted 2 times total
    And the pipeline should continue

  Scenario: Stage failure after all retries applies on_fail strategy
    Given a pipeline with max_retries: 2
    And a stage with on_fail: fail that fails on all attempts
    When the stage runs
    Then the stage should be attempted 3 times total
    And the outcome should be FAIL

  Scenario: Explicit outcome signals are not retried
    Given a prompt stage that outputs "FAIL\nReason"
    When the stage runs
    Then no retries are attempted
    And the outcome is FAIL immediately

  # Lifecycle hooks

  Scenario: on_succeed runs after all stages pass, before ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_succeed hooks
    When all stages pass
    Then on_succeed hooks run
    And then ticket "ko-a001" is closed

  Scenario: on_close runs after ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_close hooks
    When "ko build ko-a001" succeeds
    Then ticket "ko-a001" is closed first
    And then on_close hooks run

  Scenario: Hook commands receive TICKET_ID and CHANGED_FILES
    Given a pipeline with on_succeed: "echo ${TICKET_ID} ${CHANGED_FILES}"
    When on_succeed runs for ticket "ko-a001"
    Then TICKET_ID is set to "ko-a001"
    And CHANGED_FILES contains files modified during the build

  Scenario: on_succeed failure prevents ticket close
    Given a ticket "ko-a001" with status "open"
    And a pipeline with an on_succeed hook that fails
    When "ko build ko-a001" runs
    Then ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "on_succeed failed"

  # CHANGED_FILES tracking

  Scenario: CHANGED_FILES captures files modified during build
    Given a ticket "ko-a001" with status "open"
    And a pipeline where the implement stage modifies "foo.go" and creates "bar.go"
    When the build completes
    Then CHANGED_FILES should contain "foo.go" and "bar.go"

  # Decomposition depth guard

  Scenario: Decomposition is denied at max depth
    Given a ticket "ko-a001.b002" at depth 1
    And a pipeline with max_depth: 1
    And a prompt stage that outputs "DECOMPOSE\n- Subtask"
    When the stage runs
    Then ticket "ko-a001.b002" should have status "blocked"
    And ticket "ko-a001.b002" should have a note containing "max decomposition depth"
    And no child tickets should have been created

  Scenario: Decomposition is allowed below max depth
    Given a ticket "ko-a001" at depth 0
    And a pipeline with max_depth: 2
    And a prompt stage that outputs "DECOMPOSE\n- Subtask one\n- Subtask two"
    When the stage runs
    Then 2 child tickets should exist with IDs starting with "ko-a001."

  # External ask limit

  Scenario: Build can create at most one external ask per run
    Given a ticket "ko-a001" with status "open"
    And a pipeline where a stage requests 2 external asks
    When "ko build ko-a001" runs
    Then exactly 1 external ask should have been created
    And ticket "ko-a001" should have a note about throttled external asks

  # Build artifacts

  Scenario: Build creates an artifact directory
    Given a ticket "ko-a001" with status "open"
    When I run "ko build ko-a001"
    Then a build directory should exist under ".ko/builds/"
    And it should contain the ticket snapshot
    And it should contain output from each stage

  # Loop safety invariant

  Scenario: No outcome leaves ticket unchanged on ready queue
    Given a ticket "ko-a001" with status "open"
    When "ko build ko-a001" completes with any outcome
    Then ticket "ko-a001" should not appear in "ko ready" output

  # Discretion levels

  Scenario: Discretion level is passed to prompt stages
    Given a pipeline with discretion: high
    And a prompt stage
    When the stage runs
    Then the prompt context includes the discretion level "high"
    And the discretion guidance text for "high" is included
