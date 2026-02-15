Feature: Directory Resolution
  As a user
  I want ko to find .tickets by walking parent directories
  So that I can run commands from any subdirectory of my project

  Background:
    Given a clean tickets directory

  Scenario: Find tickets in current directory
    Given a ticket exists with ID "test-0001" and title "Test ticket"
    When I run "ko ls"
    Then the command should succeed
    And the output should contain "test-0001"

  Scenario: Find tickets in parent directory
    Given a ticket exists with ID "test-0001" and title "Test ticket"
    And I am in subdirectory "src/components"
    When I run "ko ls"
    Then the command should succeed
    And the output should contain "test-0001"

  Scenario: Find tickets in grandparent directory
    Given a ticket exists with ID "test-0001" and title "Test ticket"
    And I am in subdirectory "src/components/ui"
    When I run "ko ls"
    Then the command should succeed
    And the output should contain "test-0001"

  Scenario: Error when no tickets directory found
    Given the tickets directory does not exist
    When I run "ko ls"
    Then the command should fail
    And the error should contain "no .tickets directory found"

  Scenario: TICKETS_DIR env var takes priority
    Given a ticket exists with ID "parent-0001" and title "Parent ticket"
    And a separate tickets directory exists at "other-tickets" with ticket "other-0001"
    When I run "ko ls" with TICKETS_DIR set to "other-tickets"
    Then the command should succeed
    And the output should contain "other-0001"
    And the output should not contain "parent-0001"

  Scenario: Help works without tickets directory
    Given the tickets directory does not exist
    When I run "ko help"
    Then the command should succeed
