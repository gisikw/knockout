Feature: Ticket Hierarchy
  As a user
  I want tickets with hierarchical IDs
  So that decomposition depth is visible and enforceable

  Background:
    Given a clean tickets directory

  Scenario: Child ticket ID encodes parent
    Given a ticket exists with ID "ko-a001" and title "Parent ticket"
    When I run "ko create 'Child ticket' --parent ko-a001"
    Then the command should succeed
    And the created ticket ID should start with "ko-a001."
    And the created ticket ID should have 1 dot

  Scenario: Grandchild ticket ID encodes full lineage
    Given a ticket exists with ID "ko-a001" and title "Root"
    And a ticket exists with ID "ko-a001.b002" and title "Child" with parent "ko-a001"
    When I run "ko create 'Grandchild' --parent ko-a001.b002"
    Then the command should succeed
    And the created ticket ID should start with "ko-a001.b002."
    And the created ticket ID should have 2 dots

  Scenario: Depth is readable from ID
    Given a ticket exists with ID "ko-a001" and title "Root"
    And a ticket exists with ID "ko-a001.b002" and title "Child" with parent "ko-a001"
    And a ticket exists with ID "ko-a001.b002.c003" and title "Grandchild" with parent "ko-a001.b002"
    Then ticket "ko-a001" should have depth 0
    And ticket "ko-a001.b002" should have depth 1
    And ticket "ko-a001.b002.c003" should have depth 2

  Scenario: Show displays children section
    Given a ticket exists with ID "ko-a001" and title "Parent"
    And a ticket exists with ID "ko-a001.b002" and title "Child" with parent "ko-a001"
    When I run "ko show ko-a001"
    Then the command should succeed
    And the output should contain "## Children"
    And the output should contain "ko-a001.b002"

  Scenario: Parent must exist
    When I run "ko create 'Orphan' --parent nonexistent"
    Then the command should fail
    And the error should contain "ticket 'nonexistent' not found"

  Scenario: Child ticket file uses full hierarchical ID
    Given a ticket exists with ID "ko-a001" and title "Parent"
    When I run "ko create 'Child' --parent ko-a001"
    Then a ticket file should exist matching pattern "ko-a001.*.md"
