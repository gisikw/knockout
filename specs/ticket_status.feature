Feature: Ticket Status Management
  As a user
  I want to change ticket statuses
  So that I can track progress on tasks

  Background:
    Given a clean tickets directory
    And a ticket exists with ID "test-0001" and title "Test ticket"

  Scenario: Set status to in_progress
    When I run "ko status test-0001 in_progress"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Set status to closed
    When I run "ko status test-0001 closed"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "closed"

  Scenario: Set status to blocked
    When I run "ko status test-0001 blocked"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"

  Scenario: Set status to open from closed
    Given ticket "test-0001" has status "closed"
    When I run "ko status test-0001 open"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "open"

  Scenario: Start command sets status to in_progress
    When I run "ko start test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Close command sets status to closed
    When I run "ko close test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "closed"

  Scenario: Open command sets status to open
    Given ticket "test-0001" has status "closed"
    When I run "ko open test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "open"

  Scenario: Block command sets status to blocked
    When I run "ko block test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"

  Scenario: Block with valid questions JSON
    When I run "ko block test-0001 --questions '[{\"id\":\"q1\",\"question\":\"Which approach?\",\"options\":[{\"label\":\"Option A\",\"value\":\"a\"},{\"label\":\"Option B\",\"value\":\"b\"}]}]'"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"
    And ticket "test-0001" should have plan-questions with 1 question

  Scenario: Block without questions flag
    When I run "ko block test-0001"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"
    And ticket "test-0001" should not have plan-questions

  Scenario: Block with invalid questions JSON
    When I run "ko block test-0001 --questions 'not valid json'"
    Then the command should fail
    And the error should contain "invalid JSON"

  Scenario: Block with questions missing required field id
    When I run "ko block test-0001 --questions '[{\"question\":\"Test?\",\"options\":[{\"label\":\"A\",\"value\":\"a\"}]}]'"
    Then the command should fail
    And the error should contain "missing required field 'id'"

  Scenario: Block with questions missing required field question
    When I run "ko block test-0001 --questions '[{\"id\":\"q1\",\"options\":[{\"label\":\"A\",\"value\":\"a\"}]}]'"
    Then the command should fail
    And the error should contain "missing required field 'question'"

  Scenario: Block with questions missing options
    When I run "ko block test-0001 --questions '[{\"id\":\"q1\",\"question\":\"Test?\",\"options\":[]}]'"
    Then the command should fail
    And the error should contain "missing required field 'options'"

  Scenario: Block with option missing label
    When I run "ko block test-0001 --questions '[{\"id\":\"q1\",\"question\":\"Test?\",\"options\":[{\"value\":\"a\"}]}]'"
    Then the command should fail
    And the error should contain "missing required field 'label'"

  Scenario: Block with option missing value
    When I run "ko block test-0001 --questions '[{\"id\":\"q1\",\"question\":\"Test?\",\"options\":[{\"label\":\"A\"}]}]'"
    Then the command should fail
    And the error should contain "missing required field 'value'"

  Scenario: Valid statuses are a closed set
    When I run "ko status test-0001 invalid"
    Then the command should fail
    And the error should contain "invalid status"
    And the error should contain "captured routed open in_progress closed blocked"

  Scenario: Status of non-existent ticket
    When I run "ko status nonexistent open"
    Then the command should fail
    And the error should contain "ticket 'nonexistent' not found"

  Scenario: Status command with partial ID
    When I run "ko status 0001 in_progress"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "in_progress"

  Scenario: Blocked tickets are excluded from ready
    Given ticket "test-0001" has status "blocked"
    When I run "ko ready"
    Then the output should not contain "test-0001"

  Scenario: Captured tickets are excluded from ready
    Given ticket "test-0001" has status "captured"
    When I run "ko ready"
    Then the output should not contain "test-0001"

  Scenario: Routed tickets are excluded from ready
    Given ticket "test-0001" has status "routed"
    When I run "ko ready"
    Then the output should not contain "test-0001"

  Scenario: Answer partial plan questions
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Tabs or spaces?","options":[{"label":"Tabs","value":"tabs"},{"label":"Spaces","value":"spaces"}]},{"id":"q2","question":"Fix manually or with script?","options":[{"label":"Manual","value":"manual"},{"label":"Script","value":"script"}]}]'
    When I run "ko answer test-0001 '{\"q1\":\"Spaces, 2-wide\"}'"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "blocked"
    And ticket "test-0001" should have plan-questions with 1 question
    And ticket "test-0001" body should contain "Plan answer (q1): Tabs or spaces? → Spaces, 2-wide"

  Scenario: Answer all plan questions unblocks ticket
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Tabs or spaces?","options":[{"label":"Tabs","value":"tabs"},{"label":"Spaces","value":"spaces"}]},{"id":"q2","question":"Fix manually or with script?","options":[{"label":"Manual","value":"manual"},{"label":"Script","value":"script"}]}]'
    When I run "ko answer test-0001 '{\"q1\":\"Spaces, 2-wide\",\"q2\":\"I will fix manually\"}'"
    Then the command should succeed
    And ticket "test-0001" should have field "status" with value "open"
    And ticket "test-0001" should not have plan-questions
    And ticket "test-0001" body should contain "Plan answer (q1): Tabs or spaces? → Spaces, 2-wide"
    And ticket "test-0001" body should contain "Plan answer (q2): Fix manually or with script? → I will fix manually"

  Scenario: Answer command with invalid JSON
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Test?","options":[{"label":"Yes","value":"yes"}]}]'
    When I run "ko answer test-0001 'invalid json'"
    Then the command should fail
    And the error should contain "invalid JSON"

  Scenario: Answer command with nonexistent question ID
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Test?","options":[{"label":"Yes","value":"yes"}]}]'
    When I run "ko answer test-0001 '{\"q99\":\"answer\"}'"
    Then the command should fail
    And the error should contain "question ID q99 not found"

  Scenario: Answer command on ticket with no plan-questions
    Given ticket "test-0001" has status "open"
    When I run "ko answer test-0001 '{\"q1\":\"answer\"}'"
    Then the command should fail
    And the error should contain "has no plan-questions"

  Scenario: Show plan questions as JSON
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Tabs or spaces?","options":[{"label":"Tabs","value":"tabs"},{"label":"Spaces","value":"spaces"}]},{"id":"q2","question":"Fix manually or with script?","options":[{"label":"Manual","value":"manual"},{"label":"Script","value":"script"}]}]'
    When I run "ko questions test-0001"
    Then the command should succeed
    And the output should be valid JSON
    And the JSON should contain 2 questions
    And the JSON question 0 should have id "q1"
    And the JSON question 1 should have id "q2"

  Scenario: Show plan questions for ticket with no questions
    Given ticket "test-0001" has status "open"
    When I run "ko questions test-0001"
    Then the command should succeed
    And the output should be "[]"

  Scenario: Show plan questions with partial ID
    Given ticket "test-0001" is blocked with plan-questions '[{"id":"q1","question":"Test?","options":[{"label":"Yes","value":"yes"}]}]'
    When I run "ko questions 0001"
    Then the command should succeed
    And the output should be valid JSON
    And the JSON should contain 1 question

  Scenario: Show plan questions for nonexistent ticket
    When I run "ko questions nonexistent"
    Then the command should fail
    And the error should contain "not found"
