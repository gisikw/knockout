Feature: Ticket Notes
  As a user
  I want to add notes to tickets
  So that I can track progress and updates

  Background:
    Given a clean tickets directory
    And a ticket exists with ID "note-0001" and title "Test ticket"

  Scenario: Add a note to ticket
    When I run "ko add-note note-0001 'This is my note'"
    Then the command should succeed
    And ticket "note-0001" should contain "## Notes"
    And ticket "note-0001" should contain "This is my note"

  Scenario: Note has timestamp
    When I run "ko add-note note-0001 'Timestamped note'"
    Then the command should succeed
    And ticket "note-0001" should contain a timestamp in notes

  Scenario: Add multiple notes
    When I run "ko add-note note-0001 'First note'"
    And I run "ko add-note note-0001 'Second note'"
    Then ticket "note-0001" should contain "First note"
    And ticket "note-0001" should contain "Second note"

  Scenario: Add note to non-existent ticket
    When I run "ko add-note nonexistent 'My note'"
    Then the command should fail
    And the error should contain "not found"

  Scenario: Add note with partial ID
    When I run "ko add-note 0001 'Partial ID note'"
    Then the command should succeed

  Scenario: Add multiline note via stdin
    When I pipe "First line\nSecond line\nThird line" to "ko add-note note-0001"
    Then the command should succeed
    And ticket "note-0001" should contain "First line"
    And ticket "note-0001" should contain "Second line"
    And ticket "note-0001" should contain "Third line"

  Scenario: Add note via heredoc
    When I run "ko add-note note-0001" with heredoc input
    Then the command should succeed
    And ticket "note-0001" should preserve multiline formatting

  Scenario: Stdin takes precedence over command-line args
    When I pipe "From stdin" to "ko add-note note-0001 'From args'"
    Then the command should succeed
    And ticket "note-0001" should contain "From stdin"
    And ticket "note-0001" should not contain "From args"

  Scenario: Empty stdin falls back to args
    When I run "ko add-note note-0001 'From command line'" with no stdin
    Then the command should succeed
    And ticket "note-0001" should contain "From command line"
