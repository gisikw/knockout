Feature: Project Registry
  As a user
  I want a central registry of projects
  So that routing and cross-project deps can be resolved

  Scenario: Registry is a config file
    Given a registry file at "~/.config/knockout/projects.yml"
    When I run "ko projects"
    Then the command should succeed
    And the output should list registered projects

  Scenario: Registry has a default project
    Given a registry with default project "exo"
    When I run "ko add 'Unroutable thing #nonexistent'"
    Then a ticket should exist in "exo" with tag "nonexistent"

  Scenario: Missing registry file shows helpful error
    Given no registry file exists
    When I run "ko projects"
    Then the command should fail
    And the error should contain "no project registry found"
    And the error should contain "knockout/projects.yml"

  # Cross-project dep resolution

  Scenario: Ready checks cross-project deps only when local queue empty
    Given project "alpha" has no locally ready tickets
    And project "alpha" has a ticket blocked on "beta-0001" in project "beta"
    And "beta-0001" is closed in project "beta"
    When I run "ko ready" in project "alpha"
    Then the output should contain the cross-project unblocked ticket

  Scenario: Local ready tickets take priority over cross-project checks
    Given project "alpha" has a locally ready ticket "alpha-0001"
    And project "alpha" has a ticket blocked on "beta-0001" in project "beta"
    And "beta-0001" is closed in project "beta"
    When I run "ko ready" in project "alpha"
    Then the output should contain "alpha-0001"
    And cross-project deps should not have been checked

  Scenario: Cross-project dep check short-circuits on first match
    Given project "alpha" has no locally ready tickets
    And project "alpha" has 10 tickets blocked on cross-project deps
    And 3 of those cross-project deps are resolved
    When I run "ko ready" in project "alpha"
    Then the output should contain exactly 1 ticket
