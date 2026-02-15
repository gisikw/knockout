Feature: Ticket Status Management
  As a user
  I want to change ticket statuses
  So that I can track progress on tasks

  Background:
    Given a clean tickets directory
    And a ticket exists with ID "test-0001" and title "Test ticket"

  Scenario: Set status to in_progress
    When I run "ko status test-0001 in_progress"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Set status to closed
    When I run "ko status test-0001 closed"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "closed"

  Scenario: Set status to blocked
    When I run "ko status test-0001 blocked"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"

  Scenario: Set status to open from closed
    Given ticket "test-0001" has status "closed"
    When I run "ko status test-0001 open"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "open"

  Scenario: Start command sets status to in_progress
    When I run "ko start test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Close command sets status to closed
    When I run "ko close test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "closed"

  Scenario: Reopen command sets status to open
    Given ticket "test-0001" has status "closed"
    When I run "ko reopen test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "open"

  Scenario: Block command sets status to blocked
    When I run "ko block test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"

  Scenario: Valid statuses are a closed set
    When I run "ko status test-0001 invalid"
    Then the command should fail
    And the error should contain "invalid status"
    And the error should contain "captured routed open in_progress closed blocked"

  Scenario: Status of non-existent ticket
    When I run "ko status nonexistent open"
    Then the command should fail
    And the error should contain "ticket 'nonexistent' not found"

  Scenario: Status command with partial ID
    When I run "ko status 0001 in_progress"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Blocked tickets are excluded from ready
    Given ticket "test-0001" has status "blocked"
    When I run "ko ready"
    Then the output should not contain "test-0001"

  Scenario: Captured tickets are excluded from ready
    Given ticket "test-0001" has status "captured"
    When I run "ko ready"
    Then the output should not contain "test-0001"

  Scenario: Routed tickets are excluded from ready
    Given ticket "test-0001" has status "routed"
    When I run "ko ready"
    Then the output should not contain "test-0001"
