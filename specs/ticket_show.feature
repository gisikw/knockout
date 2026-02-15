Feature: Ticket Show
  As a user
  I want to view ticket details
  So that I can see full information about a ticket

  Background:
    Given a clean tickets directory

  Scenario: Show displays ticket content
    Given a ticket exists with ID "show-001" and title "Test ticket"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should contain "id: show-001"
    And the output should contain "# Test ticket"

  Scenario: Show displays all frontmatter fields
    Given a ticket exists with ID "show-001" and title "Full ticket"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should contain "status: open"
    And the output should contain "deps: []"
    And the output should contain "links: []"
    And the output should contain "type: task"
    And the output should contain "priority: 2"

  Scenario: Show displays blockers section when deps are open
    Given a ticket exists with ID "show-001" and title "Blocked ticket"
    And a ticket exists with ID "show-002" and title "Blocker ticket"
    And ticket "show-001" depends on "show-002"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should contain "## Blockers"
    And the output should contain "show-002"

  Scenario: Show hides blockers section when all deps closed
    Given a ticket exists with ID "show-001" and title "Unblocked ticket"
    And a ticket exists with ID "show-002" and title "Closed blocker"
    And ticket "show-001" depends on "show-002"
    And ticket "show-002" has status "closed"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should not contain "## Blockers"

  Scenario: Show displays blocking section
    Given a ticket exists with ID "show-001" and title "Blocker"
    And a ticket exists with ID "show-002" and title "Blocked"
    And ticket "show-002" depends on "show-001"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should contain "## Blocking"
    And the output should contain "show-002"

  Scenario: Show displays children section
    Given a ticket exists with ID "show-001" and title "Parent"
    And a ticket exists with ID "show-001.a002" and title "Child" with parent "show-001"
    When I run "ko show show-001"
    Then the command should succeed
    And the output should contain "## Children"
    And the output should contain "show-001.a002"

  Scenario: Show non-existent ticket
    When I run "ko show nonexistent"
    Then the command should fail
    And the error should contain "not found"

  Scenario: Show with partial ID
    Given a ticket exists with ID "show-001" and title "Test ticket"
    When I run "ko show 001"
    Then the command should succeed
    And the output should contain "id: show-001"
