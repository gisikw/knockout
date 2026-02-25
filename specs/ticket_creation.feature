Feature: Ticket Creation
  As a user
  I want to create tickets with various options
  So that I can track tasks in my project

  Background:
    Given a clean tickets directory

  Scenario: Create a basic ticket with title
    When I run "ko add 'My first ticket'"
    Then the command should succeed
    And the output should match a ticket ID pattern
    And a ticket file should exist with title "My first ticket"

  Scenario: Create a ticket with default title
    When I run "ko add"
    Then the command should succeed
    And the output should match a ticket ID pattern
    And a ticket file should exist with title "Untitled"

  Scenario: Create a ticket with description using -d flag
    When I run "ko add 'Test ticket' -d 'This is the description'"
    Then the command should succeed
    And the created ticket should contain "This is the description"

  Scenario: Create a ticket with description as second positional argument
    When I run "ko add 'Test ticket' 'Description from second arg'"
    Then the command should succeed
    And the created ticket should contain "Description from second arg"

  Scenario: Create a ticket with description from stdin
    When I pipe "Description from stdin" to "ko add 'Test ticket'"
    Then the command should succeed
    And the created ticket should contain "Description from stdin"

  Scenario: Second positional argument takes priority over -d flag
    When I run "ko add 'Test ticket' 'From arg' -d 'From flag'"
    Then the command should succeed
    And the created ticket should contain "From arg"
    And the created ticket should not contain "From flag"

  Scenario: Stdin takes priority over second positional argument
    When I pipe "From stdin" to "ko add 'Test ticket' 'From arg'"
    Then the command should succeed
    And the created ticket should contain "From stdin"
    And the created ticket should not contain "From arg"

  Scenario: Stdin takes priority over -d flag
    When I pipe "From stdin" to "ko add 'Test ticket' -d 'From flag'"
    Then the command should succeed
    And the created ticket should contain "From stdin"
    And the created ticket should not contain "From flag"

  Scenario: Create a ticket with type
    When I run "ko add 'Bug ticket' -t bug"
    Then the command should succeed
    And the created ticket should have field "type" with value "bug"

  Scenario: Create a ticket with priority
    When I run "ko add 'High priority' -p 0"
    Then the command should succeed
    And the created ticket should have field "priority" with value "0"

  Scenario: Create a ticket with shorthand priority flag
    When I run "ko add 'Shorthand priority' -p1"
    Then the command should succeed
    And the created ticket should have field "priority" with value "1"

  Scenario: Create a ticket with assignee
    When I run "ko add 'Assigned ticket' -a 'John Doe'"
    Then the command should succeed
    And the created ticket should have field "assignee" with value "John Doe"

  Scenario: Create a ticket with external reference
    When I run "ko add 'External ticket' --external-ref 'JIRA-123'"
    Then the command should succeed
    And the created ticket should have field "external-ref" with value "JIRA-123"

  Scenario: Create a ticket with parent
    Given a ticket exists with ID "ko-a001" and title "Parent ticket"
    When I run "ko add 'Child ticket' --parent ko-a001"
    Then the command should succeed
    And the created ticket ID should start with "ko-a001."
    And the created ticket should have field "parent" with value "ko-a001"

  Scenario: Create a ticket with design notes
    When I run "ko add 'Design ticket' --design 'Use microservices'"
    Then the command should succeed
    And the created ticket should contain "## Design"
    And the created ticket should contain "Use microservices"

  Scenario: Create a ticket with acceptance criteria
    When I run "ko add 'Story ticket' --acceptance 'Should pass all tests'"
    Then the command should succeed
    And the created ticket should contain "## Acceptance Criteria"
    And the created ticket should contain "Should pass all tests"

  Scenario: Create a ticket with tags
    When I run "ko add 'Tagged ticket' --tags 'ui,backend'"
    Then the command should succeed
    And the created ticket should have tags "ui" and "backend"

  Scenario: Ticket has default status open
    When I run "ko add 'New ticket'"
    Then the command should succeed
    And the created ticket should have field "status" with value "open"

  Scenario: Ticket has default priority 2
    When I run "ko add 'Normal priority'"
    Then the command should succeed
    And the created ticket should have field "priority" with value "2"

  Scenario: Ticket has default type task
    When I run "ko add 'Default type'"
    Then the command should succeed
    And the created ticket should have field "type" with value "task"

  Scenario: Ticket has empty deps by default
    When I run "ko add 'No deps'"
    Then the command should succeed
    And the created ticket should have field "deps" with value "[]"

  Scenario: Ticket has empty links by default
    When I run "ko add 'No links'"
    Then the command should succeed
    And the created ticket should have field "links" with value "[]"

  Scenario: Ticket has created timestamp
    When I run "ko add 'Timestamped'"
    Then the command should succeed
    And the created ticket should have a valid created timestamp

  Scenario: Tickets directory created on demand
    Given the tickets directory does not exist
    When I run "ko add 'First ticket'"
    Then the command should succeed
    And the tickets directory should exist

  Scenario: Prefix matches existing tickets
    Given a ticket exists with ID "myproj-a001" and title "Existing"
    When I run "ko add 'Another ticket'"
    Then the command should succeed
    And the created ticket ID should start with "myproj-"

  Scenario: Prefix derived from directory name when no tickets exist
    Given a clean tickets directory in project "my-cool-project"
    When I run "ko add 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "mcp-"

  Scenario: Single-segment directory uses first three characters
    Given a clean tickets directory in project "exocortex"
    When I run "ko add 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "exo-"

  Scenario: Underscore-separated directory name uses initials
    Given a clean tickets directory in project "fort_nix"
    When I run "ko add 'First ticket'"
    Then the command should succeed
    And the created ticket ID should start with "fn-"
