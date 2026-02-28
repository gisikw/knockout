Feature: Ticket Snooze Field
  As a user
  I want to set a snooze date on tickets
  So that I can defer work until a later date

  Background:
    Given a clean tickets directory

  Scenario: Create a ticket with a snooze date
    When I run "ko add 'Deferred task' --snooze 2026-05-01"
    Then the command should succeed
    And the created ticket frontmatter should contain "snooze: 2026-05-01"

  Scenario: Update a ticket to set a snooze date
    Given a ticket exists with ID "ko-a001" and title "Task to snooze"
    When I run "ko update ko-a001 --snooze 2026-05-01"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should contain "snooze: 2026-05-01"

  Scenario: Invalid snooze date format is rejected on create
    When I run "ko add 'Bad snooze' --snooze not-a-date"
    Then the command should fail with a non-zero exit code

  Scenario: Invalid snooze date format is rejected on update
    Given a ticket exists with ID "ko-a001" and title "Task"
    When I run "ko update ko-a001 --snooze bad"
    Then the command should fail with a non-zero exit code

  Scenario: Ticket without snooze has no snooze field in frontmatter
    When I run "ko add 'No snooze'"
    Then the command should succeed
    And the created ticket frontmatter should not contain "snooze:"

  Scenario: Snooze shorthand sets snooze field
    Given a ticket exists with ID "ko-a001" and title "Task to snooze"
    When I run "ko snooze ko-a001 2026-05-01"
    Then the command should succeed
    And the ticket "ko-a001" frontmatter should contain "snooze: 2026-05-01"

  Scenario: Snooze shorthand rejects missing date argument
    Given a ticket exists with ID "ko-a001" and title "Task"
    When I run "ko snooze ko-a001"
    Then the command should fail with a non-zero exit code

  Scenario: Snooze shorthand rejects invalid date format
    Given a ticket exists with ID "ko-a001" and title "Task"
    When I run "ko snooze ko-a001 not-a-date"
    Then the command should fail with a non-zero exit code
