Feature: Ticket ID Resolution
  As a user
  I want to use partial ticket IDs
  So that I can work faster without typing full IDs

  Background:
    Given a clean tickets directory

  Scenario: Exact ID match
    Given a ticket exists with ID "abc-1234" and title "Test ticket"
    When I run "ko show abc-1234"
    Then the command should succeed
    And the output should contain "id: abc-1234"

  Scenario: Partial ID match by suffix
    Given a ticket exists with ID "abc-1234" and title "Test ticket"
    When I run "ko show 1234"
    Then the command should succeed
    And the output should contain "id: abc-1234"

  Scenario: Partial ID match by prefix
    Given a ticket exists with ID "abc-1234" and title "Test ticket"
    When I run "ko show abc"
    Then the command should succeed
    And the output should contain "id: abc-1234"

  Scenario: Partial ID match by substring
    Given a ticket exists with ID "abc-1234" and title "Test ticket"
    When I run "ko show c-12"
    Then the command should succeed
    And the output should contain "id: abc-1234"

  Scenario: Ambiguous ID error
    Given a ticket exists with ID "abc-1234" and title "First ticket"
    And a ticket exists with ID "abc-5678" and title "Second ticket"
    When I run "ko show abc"
    Then the command should fail
    And the error should contain "ambiguous"

  Scenario: Non-existent ID error
    When I run "ko show nonexistent"
    Then the command should fail
    And the error should contain "not found"

  Scenario: Exact match takes precedence over partial
    Given a ticket exists with ID "abc" and title "Short ID ticket"
    And a ticket exists with ID "abc-1234" and title "Long ID ticket"
    When I run "ko show abc"
    Then the command should succeed
    And the output should contain "id: abc"

  Scenario: Partial ID works with hierarchical IDs
    Given a ticket exists with ID "ko-a001" and title "Parent"
    And a ticket exists with ID "ko-a001.b002" and title "Child"
    When I run "ko show b002"
    Then the command should succeed
    And the output should contain "id: ko-a001.b002"

  Scenario: ID resolution works across commands
    Given a ticket exists with ID "test-9999" and title "Test ticket"
    When I run "ko status 9999 in_progress"
    Then the command should succeed
    And ticket "test-9999" should have field "status" with value "in_progress"
