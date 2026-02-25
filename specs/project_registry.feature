Feature: Project Registry
  As a user
  I want a central registry of projects
  So that routing and cross-project deps can be resolved

  # Registration

  Scenario: Register a project with a tag
    Given I am in a project directory "/tmp/test-projects/fort-nix"
    When I run "ko project set #fort-nix"
    Then the command should succeed
    And the registry at "~/.config/knockout/projects.yml" should contain project "fort-nix"
    And the registered path for "fort-nix" should be "/tmp/test-projects/fort-nix"

  Scenario: Register strips the hash from the tag
    Given I am in a project directory "/tmp/test-projects/exo"
    When I run "ko project set #exo"
    Then the registered project name should be "exo" not "#exo"

  Scenario: Register requires a tag argument
    When I run "ko project set"
    Then the command should fail
    And the error should contain "#tag argument required"

  Scenario: Register overwrites existing entry for same tag (upsert)
    Given a registry with project "exo" at "/old/path"
    And I am in a project directory "/new/path"
    When I run "ko project set #exo"
    Then the command should succeed
    And the registered path for "exo" should be "/new/path"

  Scenario: Register creates registry file if it does not exist
    Given no registry file exists
    And I am in a project directory "/tmp/test-projects/fort-nix"
    When I run "ko project set #fort-nix"
    Then the command should succeed
    And the registry file should exist

  Scenario: Set project with prefix
    Given I am in a project directory "/tmp/test-projects/fort-nix"
    When I run "ko project set #fort-nix --prefix=nix"
    Then the command should succeed
    And the prefix "nix" should be stored for project "fort-nix"
    And the .ko/config.yaml should contain prefix "nix"

  # Default project

  Scenario: Set default project during registration
    Given I am in a project directory "/tmp/test-projects/exo"
    When I run "ko project set #exo --default"
    Then the command should succeed
    And the registry default should be "exo"

  Scenario: Set project as default after initial registration
    Given a registry with project "exo" at "/tmp/test-projects/exo"
    When I run "ko project set #exo --default"
    Then the command should succeed
    And the registry default should be "exo"

  Scenario: Upsert can update prefix and set default in one command
    Given a registry with project "exo" at "/tmp/test-projects/exo"
    When I run "ko project set #exo --prefix=new --default"
    Then the command should succeed
    And the prefix "new" should be stored for project "exo"
    And the registry default should be "exo"

  # Listing

  Scenario: List all registered projects
    Given a registry file at "~/.config/knockout/projects.yml"
    When I run "ko project ls"
    Then the command should succeed
    And the output should list registered projects

  Scenario: List shows default project with marker
    Given a registry with default project "exo"
    When I run "ko project ls"
    Then the command should succeed
    And the output should show "exo" with an asterisk marker

  Scenario: List shows no projects when registry is empty
    Given an empty registry
    When I run "ko project ls"
    Then the command should succeed
    And the output should contain "no projects registered"

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
