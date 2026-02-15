Feature: Ticket Creation
  As a user
  I want to create tickets with various options
  So that I can track tasks in my project

  Background:
    Given a clean tickets directory

  Scenario: Create a basic ticket with title
    When I run "ko create 'My first ticket'"
    Then the command should succeed
    And the output should match a ticket ID pattern
    And a ticket file should exist with title "My first ticket"

  Scenario: Create a ticket with default title
    When I run "ko create"
    Then the command should succeed
    And the output should match a ticket ID pattern
    And a ticket file should exist with title "Untitled"

  Scenario: Create a ticket with description
    When I run "ko create 'Test ticket' -d 'This is the description'"
    Then the command should succeed
    And the created ticket should contain "This is the description"

  Scenario: Create a ticket with type
    When I run "ko create 'Bug ticket' -t bug"
    Then the command should succeed
    And the created ticket should have field "type" with value "bug"

  Scenario: Create a ticket with priority
    When I run "ko create 'High priority' -p 0"
    Then the command should succeed
    And the created ticket should have field "priority" with value "0"

  Scenario: Create a ticket with assignee
    When I run "ko create 'Assigned ticket' -a 'John Doe'"
    Then the command should succeed
    And the created ticket should have field "assignee" with value "John Doe"

  Scenario: Create a ticket with external reference
    When I run "ko create 'External ticket' --external-ref 'JIRA-123'"
    Then the command should succeed
    And the created ticket should have field "external-ref" with value "JIRA-123"

  Scenario: Create a ticket with parent
    Given a ticket exists with ID "ko-a001" and title "Parent ticket"
    When I run "ko create 'Child ticket' --parent ko-a001"
    Then the command should succeed
    And the created ticket ID should start with "ko-a001."
    And the created ticket should have field "parent" with value "ko-a001"

  Scenario: Create a ticket with design notes
    When I run "ko create 'Design ticket' --design 'Use microservices'"
    Then the command should succeed
    And the created ticket should contain "## Design"
    And the created ticket should contain "Use microservices"

  Scenario: Create a ticket with acceptance criteria
    When I run "ko create 'Story ticket' --acceptance 'Should pass all tests'"
    Then the command should succeed
    And the created ticket should contain "## Acceptance Criteria"
    And the created ticket should contain "Should pass all tests"

  Scenario: Create a ticket with tags
    When I run "ko create 'Tagged ticket' --tags 'ui,backend'"
    Then the command should succeed
    And the created ticket should have tags "ui" and "backend"

  Scenario: Ticket has default status open
    When I run "ko create 'New ticket'"
    Then the command should succeed
    And the created ticket should have field "status" with value "open"

  Scenario: Ticket has default priority 2
    When I run "ko create 'Normal priority'"
    Then the command should succeed
    And the created ticket should have field "priority" with value "2"

  Scenario: Ticket has default type task
    When I run "ko create 'Default type'"
    Then the command should succeed
    And the created ticket should have field "type" with value "task"

  Scenario: Ticket has empty deps by default
    When I run "ko create 'No deps'"
    Then the command should succeed
    And the created ticket should have field "deps" with value "[]"

  Scenario: Ticket has empty links by default
    When I run "ko create 'No links'"
    Then the command should succeed
    And the created ticket should have field "links" with value "[]"

  Scenario: Ticket has created timestamp
    When I run "ko create 'Timestamped'"
    Then the command should succeed
    And the created ticket should have a valid created timestamp

  Scenario: Tickets directory created on demand
    Given the tickets directory does not exist
    When I run "ko create 'First ticket'"
    Then the command should succeed
    And the tickets directory should exist

  Scenario: Prefix matches existing tickets
    Given a ticket exists with ID "myproj-a001" and title "Existing"
    When I run "ko create 'Another ticket'"
    Then the command should succeed
    And the created ticket ID should start with "myproj-"

  Scenario: Prefix derived from directory name when no tickets exist
    Given a clean tickets directory in project "my-cool-project"
    When I run "ko create 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "mcp-"

  Scenario: Single-segment directory uses first three characters
    Given a clean tickets directory in project "exocortex"
    When I run "ko create 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "exo-"

  Scenario: Underscore-separated directory name uses initials
    Given a clean tickets directory in project "fort_nix"
    When I run "ko create 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "fn-"
