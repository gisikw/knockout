Feature: Ticket Triage Field
  As a user
  I want to set a triage note on tickets
  So that I can record triage decisions like "unblock this ticket" or "break this apart"

  Background:
    Given a clean tickets directory

  Scenario: Create a ticket with a triage note
    When I run "ko add 'Some task' --triage 'unblock this ticket'"
    Then the command should succeed
    And the created ticket frontmatter should contain "triage: unblock this ticket"

  Scenario: Update a ticket to set a triage note
    Given a ticket exists with ID "ko-a001" and title "Task to triage"
    When I run "ko update ko-a001 --triage 'break this apart'"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should contain "triage: break this apart"

  Scenario: Show displays the triage field
    Given a ticket exists with ID "ko-a001" and title "Task" and triage "unblock this ticket"
    When I run "ko show ko-a001"
    Then the command should succeed
    And the output should contain "triage: unblock this ticket"

  Scenario: Ticket without triage has no triage field in frontmatter
    When I run "ko add 'No triage'"
    Then the command should succeed
    And the created ticket frontmatter should not contain "triage:"

  Scenario: Ready excludes a ticket with a triage value set
    Given a ticket exists with ID "ko-a001" and status "open" and triage "unblock this ticket"
    When I run "ko ready"
    Then the command should succeed
    And the output should not contain "ko-a001"

  Scenario: Ready includes a ticket without a triage value
    Given a ticket exists with ID "ko-a001" and status "open"
    When I run "ko ready"
    Then the command should succeed
    And the output should contain "ko-a001"
