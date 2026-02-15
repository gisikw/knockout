Feature: Ticket Query
  As a user
  I want to query tickets as JSON
  So that I can process ticket data programmatically

  Background:
    Given a clean tickets directory

  Scenario: Query all tickets as JSONL
    Given a ticket exists with ID "query-001" and title "First ticket"
    And a ticket exists with ID "query-002" and title "Second ticket"
    When I run "ko query"
    Then the command should succeed
    And the output should be valid JSONL
    And the output should contain "query-001"
    And the output should contain "query-002"

  Scenario: Query includes all fields
    Given a ticket exists with ID "query-001" and title "Full ticket"
    When I run "ko query"
    Then the command should succeed
    And the JSONL output should have field "id"
    And the JSONL output should have field "status"
    And the JSONL output should have field "deps"
    And the JSONL output should have field "priority"

  Scenario: Query with no tickets
    When I run "ko query"
    Then the output should be empty
