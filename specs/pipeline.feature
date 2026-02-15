Feature: Build Pipeline
  As an automated system
  I want to run build pipelines against tickets
  So that the ready queue is systematically burned down

  # Outcomes — every outcome removes the ticket from ready

  Scenario: SUCCEED closes the ticket
    Given a ticket "ko-a001" with status "open"
    And a pipeline that succeeds on all stages
    When I run "ko build ko-a001"
    Then the command should succeed
    And ticket "ko-a001" should have status "closed"
    And ticket "ko-a001" should have a note containing "SUCCEED"
    And ticket "ko-a001" should have a note containing "Closed in:"

  Scenario: FAIL marks the ticket as blocked
    Given a ticket "ko-a001" with status "open"
    And a pipeline where the implement stage fails
    When I run "ko build ko-a001"
    Then the command should fail
    And ticket "ko-a001" should have status "blocked"
    And ticket "ko-a001" should have a note containing "FAIL"
    And ticket "ko-a001" should have a note containing a reason string

  Scenario: BLOCKED wires a dependency
    Given a ticket "ko-a001" with status "open"
    And a ticket "ko-b002" with status "open"
    And a pipeline where triage signals BLOCKED on "ko-b002"
    When I run "ko build ko-a001"
    Then the command should fail
    And ticket "ko-a001" should have "ko-b002" in deps
    And ticket "ko-a001" should have a note containing "BLOCKED"

  Scenario: DECOMPOSE creates children and blocks parent
    Given a ticket "ko-a001" with status "open"
    And a pipeline where triage signals DECOMPOSE with 3 subtasks
    When I run "ko build ko-a001"
    Then 3 child tickets should exist with IDs starting with "ko-a001."
    And ticket "ko-a001" should depend on all 3 children
    And ticket "ko-a001" should not appear in ready

  # Eligibility

  Scenario: Only open tickets can be built
    Given a ticket "ko-a001" with status "in_progress"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "already in progress"

  Scenario: Closed tickets cannot be built
    Given a ticket "ko-a001" with status "closed"
    And ticket "ko-a001" has a note containing "Closed in: abc1234"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "closed"
    And the error should contain "Closed in:"

  Scenario: Blocked tickets cannot be built
    Given a ticket "ko-a001" with status "blocked"
    When I run "ko build ko-a001"
    Then the command should fail
    And the error should contain "blocked"

  # Lifecycle hooks

  Scenario: on_succeed runs before ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_succeed that creates a marker file
    When I run "ko build ko-a001"
    Then the marker file should exist
    And ticket "ko-a001" should have status "closed"

  Scenario: on_close runs after ticket is closed
    Given a ticket "ko-a001" with status "open"
    And a pipeline with on_close that creates a marker file
    When I run "ko build ko-a001"
    Then the marker file should exist
    And ticket "ko-a001" should have status "closed"

  # Decomposition depth guard

  Scenario: Decomposition is denied at max depth
    Given a ticket "ko-a001.b002" at depth 1
    And max decomposition depth is 2
    And a pipeline where triage signals DECOMPOSE
    When I run "ko build ko-a001.b002"
    Then ticket "ko-a001.b002" should have status "blocked"
    And ticket "ko-a001.b002" should have a note containing "max decomposition depth"
    And no child tickets should have been created

  Scenario: Decomposition is allowed below max depth
    Given a ticket "ko-a001" at depth 0
    And max decomposition depth is 2
    And a pipeline where triage signals DECOMPOSE with 2 subtasks
    When I run "ko build ko-a001"
    Then 2 child tickets should exist with IDs starting with "ko-a001."

  # External ask limit

  Scenario: Build can create at most one external ask
    Given a ticket "ko-a001" with status "open"
    And a pipeline where implement requests 2 external asks
    When I run "ko build ko-a001"
    Then exactly 1 external ask should have been created
    And ticket "ko-a001" should have a note about throttled external asks

  # Loop safety invariant

  Scenario: No outcome leaves ticket unchanged on ready queue
    Given a ticket "ko-a001" with status "open"
    And a pipeline that produces any outcome
    When I run "ko build ko-a001"
    Then ticket "ko-a001" should not appear in "ko ready" output
