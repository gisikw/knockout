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

  Scenario: Triage with no args lists tickets with triage set
    Given a ticket exists with ID "ko-a001" and status "open" and triage "unblock this ticket"
    And a ticket exists with ID "ko-a002" and status "in_progress" and triage "break this apart"
    When I run "ko triage"
    Then the command should succeed
    And the output should contain "ko-a001"
    And the output should contain "ko-a002"
    And the output should contain "triage: unblock this ticket"
    And the output should contain "triage: break this apart"

  Scenario: Triage with no args excludes tickets without triage set
    Given a ticket exists with ID "ko-a001" and status "open" and triage "unblock this ticket"
    And a ticket exists with ID "ko-a002" and status "open"
    When I run "ko triage"
    Then the command should succeed
    And the output should contain "ko-a001"
    And the output should not contain "ko-a002"

  Scenario: Triage with --json outputs JSON with triage field
    Given a ticket exists with ID "ko-a001" and status "open" and triage "unblock this ticket"
    When I run "ko triage --json"
    Then the command should succeed
    And the output should contain "\"triage\": \"unblock this ticket\""

  Scenario: ko triage <id> <instructions> sets the triage field
    Given a ticket exists with ID "ko-a001" and title "Task to triage"
    When I run "ko triage ko-a001 break this apart"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should contain "triage: break this apart"

  Scenario: ko triage with multi-word instructions joins them
    Given a ticket exists with ID "ko-a001" and title "Task to triage"
    When I run "ko triage ko-a001 unblock this ticket now"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should contain "triage: unblock this ticket now"

  Scenario: ko triage with id but no instructions fails
    Given a ticket exists with ID "ko-a001" and title "Task to triage"
    When I run "ko triage ko-a001"
    Then the command should fail

  Scenario: ko agent triage runs triage instructions and clears the triage field
    Given a ticket exists with ID "ko-a001" and title "Fix the auth bug" and triage "unblock this ticket"
    And a pipeline config with a mock harness exists
    When I run "ko agent triage ko-a001"
    Then the command should succeed
    And the output should contain "ko-a001: triage cleared"
    And the ticket "ko-a001" frontmatter should not contain "triage:"

  Scenario: ko agent triage fails when ticket has no triage value
    Given a ticket exists with ID "ko-a001" and title "Fix the auth bug"
    And a pipeline config with a mock harness exists
    When I run "ko agent triage ko-a001"
    Then the command should fail
    And the error output should contain "has no triage value"

  Scenario: ko agent triage fails when no pipeline config exists
    Given a ticket exists with ID "ko-a001" and title "Fix the auth bug" and triage "unblock this ticket"
    When I run "ko agent triage ko-a001"
    Then the command should fail
    And the error output should contain "no config found"

  Scenario: auto_triage: true triggers triage automatically after ko add --triage
    Given a pipeline config with auto_triage: true and a mock harness exists
    When I run "ko add 'Auto task' --triage 'unblock this'"
    Then the command should succeed
    And the created ticket frontmatter should not contain "triage:"

  Scenario: auto_triage: true triggers triage automatically after ko update --triage
    Given a ticket exists with ID "ko-a001" and title "Task to triage"
    And a pipeline config with auto_triage: true and a mock harness exists
    When I run "ko update ko-a001 --triage 'break apart'"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should not contain "triage:"

  Scenario: auto_triage absent does not trigger triage automatically
    Given a pipeline config without auto_triage exists
    When I run "ko add 'Task' --triage 'unblock this'"
    Then the command should succeed
    And the created ticket frontmatter should contain "triage: unblock this"

  Scenario: auto-triage failure is non-fatal for ko add
    Given a pipeline config with auto_triage: true and a failing mock harness exists
    When I run "ko add 'Task' --triage 'unblock this'"
    Then the command should succeed
    And the created ticket frontmatter should contain "triage: unblock this"
    And the error output should contain "auto-triage for"
    And the error output should contain "failed"

  Scenario: ko triage sets triage on a cross-project ticket by ID prefix
    Given a project "fn" is registered in the registry with a ticket "fn-test"
    And the current directory is not the "fn" project
    When I run "ko triage fn-test do something"
    Then the command should succeed
    And the ticket "fn-test" in the "fn" project should have triage "do something"
