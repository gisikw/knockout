Feature: Ticket Listing
  As a user
  I want to list tickets in various ways
  So that I can see what work needs to be done

  Background:
    Given a clean tickets directory

  Scenario: List all open tickets
    Given a ticket exists with ID "list-0001" and title "First ticket"
    And a ticket exists with ID "list-0002" and title "Second ticket"
    When I run "ko ls"
    Then the command should succeed
    And the output should contain "list-0001"
    And the output should contain "list-0002"

  Scenario: List with status filter
    Given a ticket exists with ID "list-0001" and title "Open ticket"
    And a ticket exists with ID "list-0002" and title "Closed ticket"
    And ticket "list-0002" has status "closed"
    When I run "ko ls --status=open"
    Then the command should succeed
    And the output should contain "list-0001"
    And the output should not contain "list-0002"

  Scenario: List with --all flag includes closed tickets
    Given a ticket exists with ID "list-0003" and title "Open ticket"
    And a ticket exists with ID "list-0004" and title "Closed ticket"
    And ticket "list-0004" has status "closed"
    When I run "ko ls --all"
    Then the command should succeed
    And the output should contain "list-0003"
    And the output should contain "list-0004"

  Scenario: List shows dependencies
    Given a ticket exists with ID "list-0001" and title "Main ticket"
    And a ticket exists with ID "list-0002" and title "Dep ticket"
    And ticket "list-0001" depends on "list-0002"
    When I run "ko ls"
    Then the command should succeed
    And the output should contain "<- [list-0002]"

  Scenario: List with no tickets returns nothing
    When I run "ko ls"
    Then the output should be empty

  # Ready queue

  Scenario: Ready shows tickets with no deps
    Given a ticket exists with ID "ready-001" and title "Ready ticket"
    When I run "ko ready"
    Then the command should succeed
    And the output should contain "ready-001"

  Scenario: Ready shows tickets with all deps closed
    Given a ticket exists with ID "ready-001" and title "Unblocked ticket"
    And a ticket exists with ID "ready-002" and title "Dependency"
    And ticket "ready-001" depends on "ready-002"
    And ticket "ready-002" has status "closed"
    When I run "ko ready"
    Then the command should succeed
    And the output should contain "ready-001"

  Scenario: Ready excludes tickets with unclosed deps
    Given a ticket exists with ID "ready-001" and title "Blocked ticket"
    And a ticket exists with ID "ready-002" and title "Open dependency"
    And ticket "ready-001" depends on "ready-002"
    When I run "ko ready"
    Then the command should succeed
    And the output should not contain "ready-001"
    And the output should contain "ready-002"

  Scenario: Ready excludes closed tickets
    Given a ticket exists with ID "ready-001" and title "Closed ticket"
    And ticket "ready-001" has status "closed"
    When I run "ko ready"
    Then the output should not contain "ready-001"

  Scenario: Ready excludes blocked (HITL) tickets
    Given a ticket exists with ID "ready-001" and title "Needs human"
    And ticket "ready-001" has status "blocked"
    When I run "ko ready"
    Then the output should not contain "ready-001"

  Scenario: Ready excludes captured tickets
    Given a ticket exists with ID "ready-001" and title "Just captured"
    And ticket "ready-001" has status "captured"
    When I run "ko ready"
    Then the output should not contain "ready-001"

  Scenario: Ready excludes routed tickets
    Given a ticket exists with ID "ready-001" and title "Routed elsewhere"
    And ticket "ready-001" has status "routed"
    When I run "ko ready"
    Then the output should not contain "ready-001"

  Scenario: Ready sorts by priority then ID
    Given a ticket exists with ID "ready-003" and title "Low priority" with priority 3
    And a ticket exists with ID "ready-001" and title "High priority" with priority 1
    And a ticket exists with ID "ready-002" and title "Also high priority" with priority 1
    When I run "ko ready"
    Then the command should succeed
    And the output line 1 should contain "ready-001"
    And the output line 2 should contain "ready-002"
    And the output line 3 should contain "ready-003"

  # Blocked view

  Scenario: Blocked shows tickets with unclosed deps
    Given a ticket exists with ID "block-001" and title "Blocked ticket"
    And a ticket exists with ID "block-002" and title "Blocker ticket"
    And ticket "block-001" depends on "block-002"
    When I run "ko blocked"
    Then the command should succeed
    And the output should contain "block-001"
    And the output should contain "<- [block-002]"

  Scenario: Blocked excludes tickets with all deps closed
    Given a ticket exists with ID "block-001" and title "Unblocked ticket"
    And a ticket exists with ID "block-002" and title "Closed blocker"
    And ticket "block-001" depends on "block-002"
    And ticket "block-002" has status "closed"
    When I run "ko blocked"
    Then the output should not contain "block-001"

  Scenario: Blocked shows tickets with status=blocked
    Given a ticket exists with ID "block-003" and title "HITL ticket"
    And ticket "block-003" has status "blocked"
    When I run "ko blocked"
    Then the command should succeed
    And the output should contain "block-003"

  Scenario: Blocked shows block reason for status=blocked tickets
    Given a ticket exists with ID "block-004" and title "Failed ticket"
    And ticket "block-004" has status "blocked"
    And ticket "block-004" has a note "ko: FAIL at node 'actionable' — Missing required information"
    When I run "ko blocked"
    Then the command should succeed
    And the output should contain "block-004"
    And the output should contain "Missing required information"

  Scenario: Blocked with ID shows specific ticket's block reason
    Given a ticket exists with ID "block-005" and title "Specific blocked ticket"
    And ticket "block-005" has status "blocked"
    And ticket "block-005" has a note "ko: BLOCKED at node 'verify' — Test coverage insufficient"
    When I run "ko blocked block-005"
    Then the command should succeed
    And the output should contain "block-005"
    And the output should contain "Test coverage insufficient"

  Scenario: Blocked with ID shows message when no reason found
    Given a ticket exists with ID "block-006" and title "Blocked without reason"
    And ticket "block-006" has status "blocked"
    When I run "ko blocked block-006"
    Then the command should succeed
    And the output should contain "block-006: no block reason found"

  # Closed view

  Scenario: Closed shows recently closed tickets
    Given a ticket exists with ID "done-0001" and title "Done ticket"
    And ticket "done-0001" has status "closed"
    When I run "ko closed"
    Then the command should succeed
    And the output should contain "done-0001"

  Scenario: Closed respects limit
    Given a ticket exists with ID "done-0001" and title "First done"
    And a ticket exists with ID "done-0002" and title "Second done"
    And ticket "done-0001" has status "closed"
    And ticket "done-0002" has status "closed"
    When I run "ko closed --limit=1"
    Then the output line count should be 1

  # JSON output

  Scenario: List outputs JSON when --json flag is provided
    Given a ticket exists with ID "json-0001" and title "First JSON ticket"
    And a ticket exists with ID "json-0002" and title "Second JSON ticket"
    When I run "ko ls --json"
    Then the command should succeed
    And the output should be a valid JSON array
    And each JSON object should have field "id"
    And each JSON object should have field "title"
    And each JSON object should have field "status"
    And each JSON object should have field "priority"

  Scenario: Ready outputs JSON when --json flag is provided
    Given a ticket exists with ID "json-ready-001" and title "Ready JSON ticket"
    When I run "ko ready --json"
    Then the command should succeed
    And the output should be a valid JSON array
    And each JSON object should have field "id"
    And each JSON object should have field "title"

  Scenario: Blocked outputs JSON when --json flag is provided
    Given a ticket exists with ID "json-block-001" and title "Blocked JSON ticket"
    And a ticket exists with ID "json-block-002" and title "Blocker JSON ticket"
    And ticket "json-block-001" depends on "json-block-002"
    When I run "ko blocked --json"
    Then the command should succeed
    And the output should be a valid JSON array
    And each JSON object should have field "id"
    And each JSON object should have field "deps"

  Scenario: Closed outputs JSON when --json flag is provided
    Given a ticket exists with ID "json-done-001" and title "Done JSON ticket"
    And ticket "json-done-001" has status "closed"
    When I run "ko closed --json"
    Then the command should succeed
    And the output should be a valid JSON array
    And each JSON object should have field "id"
    And each JSON object should have field "status"
