Feature: Ticket Routing
  As a user
  I want to quickly capture tasks with project routing
  So that cross-project work gets to the right place without cognitive overhead

  Background:
    Given a project registry with:
      | name      | path                        |
      | fort-nix  | /tmp/test-projects/fort-nix  |
      | exo       | /tmp/test-projects/exo       |
    And "exo" is the default project
    And each project has an initialized tickets directory

  # Capture

  Scenario: Add without tag creates ticket in current project
    When I run "ko add 'Fix the flaky test'"
    Then the command should succeed
    And a ticket should exist in the current project with title "Fix the flaky test"
    And the created ticket should have field "status" with value "open"

  Scenario: Add with recognized tag routes to target project
    When I run "ko add 'Add foobar to dev-sandbox #fort-nix'"
    Then the command should succeed
    And a ticket should exist in "fort-nix" with title "Add foobar to dev-sandbox"
    And the routed ticket should have field "status" with value "routed"
    And a ticket should exist in the current project as an audit trail
    And the audit ticket should have field "status" with value "closed"

  Scenario: Add with unrecognized tag goes to default project
    When I run "ko add 'Build splash page #marketing-site'"
    Then the command should succeed
    And a ticket should exist in "exo" with title "Build splash page"
    And the created ticket should have tag "marketing-site"
    And the created ticket should have field "status" with value "captured"

  Scenario: Routed tickets are excluded from ready
    When I run "ko add 'Thing for fort-nix #fort-nix'"
    When I run "ko ready" in "fort-nix"
    Then the output should not contain "Thing for fort-nix"

  # Tag parsing

  Scenario: Tag is stripped from ticket title
    When I run "ko add 'Do the thing #fort-nix'"
    Then the routed ticket title should be "Do the thing"
    And the routed ticket title should not contain "#fort-nix"

  Scenario: Multiple tags use first as routing, rest as labels
    When I run "ko add 'Refactor auth #fort-nix #security #urgent'"
    Then a ticket should exist in "fort-nix"
    And the routed ticket should have tags "security" and "urgent"
