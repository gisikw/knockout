Feature: Agent Status JSON
  As a consumer (XO, Punchlist)
  I want to query the agent's current status and whether there is actionable work
  So that I can gate UI elements (e.g. the agent toggle) on real state

  Scenario: Not provisioned project returns actionable false
    Given a project with no pipeline config
    When I run "ko agent status --json"
    Then the JSON output should contain "provisioned": false
    And the JSON output should contain "actionable": false

  Scenario: Provisioned project with no tickets returns actionable false
    Given a project with a valid pipeline config
    And no ticket files
    When I run "ko agent status --json"
    Then the JSON output should contain "provisioned": true
    And the JSON output should contain "actionable": false

  Scenario: Provisioned project with a ready ticket returns actionable true
    Given a project with a valid pipeline config
    And a ticket with status "open" and no unresolved dependencies
    When I run "ko agent status --json"
    Then the JSON output should contain "actionable": true

  Scenario: Provisioned project with a triageable ticket returns actionable true
    Given a project with a valid pipeline config
    And a ticket with a non-empty "triage" field
    When I run "ko agent status --json"
    Then the JSON output should contain "actionable": true

  Scenario: actionable is computed regardless of whether the agent is running
    Given a project with a valid pipeline config
    And a ticket with status "open" and no unresolved dependencies
    And the agent is not running
    When I run "ko agent status --json"
    Then the JSON output should contain "running": false
    And the JSON output should contain "actionable": true
