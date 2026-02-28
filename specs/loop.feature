Feature: Build Loop
  As an operator
  I want to burn down the ready queue without human intervention
  So that all actionable work is completed automatically

  # Core loop behavior

  Scenario: Loop processes all ready tickets until queue is empty
    Given 3 tickets with status "open" and no dependencies
    And a pipeline where all stages succeed
    When I run "ko loop"
    Then all 3 tickets should be closed
    And the output should contain "3 processed"
    And the output should contain "stopped: empty"

  Scenario: Loop stops when ready queue is empty
    Given no open tickets
    When I run "ko loop"
    Then the output should contain "0 processed"
    And the output should contain "stopped: empty"

  Scenario: Loop respects dependency ordering
    Given ticket "ko-a001" with status "open"
    And ticket "ko-b002" with status "open" depending on "ko-a001"
    And a pipeline where all stages succeed
    When I run "ko loop"
    Then "ko-a001" should be built before "ko-b002"
    And both tickets should be closed

  Scenario: Failed builds do not stop the loop
    Given ticket "ko-a001" with status "open"
    And ticket "ko-b002" with status "open"
    And a pipeline where ko-a001 fails and ko-b002 succeeds
    When I run "ko loop"
    Then ticket "ko-a001" should have status "blocked"
    And ticket "ko-b002" should have status "closed"
    And the output should contain "2 processed"

  # Scope containment

  Scenario: Ticket add is blocked during loop
    Given a loop is running
    When a spawned agent runs "ko add 'New task #exo'"
    Then the command should fail
    And stderr should contain "disabled"
    And stderr should contain "runaway expansion"

  Scenario: KO_NO_CREATE is set during loop and unset after
    When I run "ko loop"
    Then KO_NO_CREATE should be set while the loop runs
    And KO_NO_CREATE should not be set after the loop completes

  # Limits

  Scenario: --max-tickets stops after N tickets
    Given 5 tickets with status "open"
    And a pipeline where all stages succeed
    When I run "ko loop --max-tickets 3"
    Then exactly 3 tickets should be processed
    And the output should contain "stopped: max_tickets"

  Scenario: --max-duration stops after elapsed time
    Given 10 tickets with status "open"
    And a pipeline where each build takes 1 second
    When I run "ko loop --max-duration 3s"
    Then at most 3 tickets should be processed
    And the output should contain "stopped: max_duration"

  # Decomposition within loop

  Scenario: Decomposed children become ready and are built in the same loop
    Given ticket "ko-a001" with status "open"
    And a pipeline where ko-a001 decomposes into 2 children that then succeed
    When I run "ko loop"
    Then ko-a001 should have been decomposed
    And both children should be closed
    And ko-a001 should be closed (deps resolved)

  # Build errors

  Scenario: Build error stops the loop
    Given ticket "ko-a001" with status "open"
    And a pipeline that causes an execution error (not an outcome signal)
    When I run "ko loop"
    Then the output should contain "stopped: build_error"

  # Loop completion hooks

  Scenario: on_loop_complete hooks run after loop finishes
    Given 2 tickets with status "open" and no dependencies
    And a pipeline with on_loop_complete hook that writes summary to a file
    When I run "ko loop"
    Then the on_loop_complete hook should have run
    And the hook should have access to LOOP_PROCESSED=2
    And the hook should have access to LOOP_SUCCEEDED=2
    And the hook should have access to LOOP_STOPPED=empty
    And the hook should have access to LOOP_RUNTIME_SECONDS

  Scenario: on_loop_complete hooks run regardless of stop reason
    Given 5 tickets with status "open"
    And a pipeline with on_loop_complete hook that writes stop reason to a file
    When I run "ko loop --max-tickets 2"
    Then the on_loop_complete hook should have run
    And the hook output should contain "LOOP_STOPPED=max_tickets"

  # Triage pre-pass

  Scenario: Loop triages all triageable tickets before processing ready tickets
    Given ticket "ko-a001" with status "open" and triage field "unblock this ticket"
    And ticket "ko-b002" with status "open" and no triage field
    And a pipeline where triage and build both succeed
    When I run "ko loop"
    Then the output should contain "triaging ko-a001"
    And both tickets should be closed

  Scenario: Loop continues building ready tickets even if triage fails for one ticket
    Given ticket "ko-a001" with status "open" and triage field set
    And a pipeline where triage fails (mock harness exits non-zero)
    And ticket "ko-b002" with status "open" and no triage field
    When I run "ko loop"
    Then the loop should log a triage failure for "ko-a001"
    And ticket "ko-b002" should still be processed and closed
