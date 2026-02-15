Feature: Ticket Links
  As a user
  I want to create symmetric links between tickets
  So that I can track related tickets

  Background:
    Given a clean tickets directory
    And a ticket exists with ID "link-0001" and title "First ticket"
    And a ticket exists with ID "link-0002" and title "Second ticket"

  Scenario: Link two tickets
    When I run "ko link link-0001 link-0002"
    Then the command should succeed
    And ticket "link-0001" should have "link-0002" in links
    And ticket "link-0002" should have "link-0001" in links

  Scenario: Link is idempotent
    Given ticket "link-0001" is linked to "link-0002"
    When I run "ko link link-0001 link-0002"
    Then the command should succeed
    And the output should contain "already exist"

  Scenario: Unlink two tickets
    Given ticket "link-0001" is linked to "link-0002"
    When I run "ko unlink link-0001 link-0002"
    Then the command should succeed
    And ticket "link-0001" should not have "link-0002" in links
    And ticket "link-0002" should not have "link-0001" in links

  Scenario: Link with non-existent ticket
    When I run "ko link link-0001 nonexistent"
    Then the command should fail
    And the error should contain "not found"
