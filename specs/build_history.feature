Feature: Build history per ticket
  Each ticket gets a persistent JSONL build history at
  `.ko/tickets/<id>.jsonl`. It is append-only and survives across
  builds and ticket close — it is the audit trail.

  Scenario: JSONL file is created on first build
    Given ticket ko-a001 exists with status open
    And a pipeline is configured
    When I run `ko agent build ko-a001`
    Then `.ko/tickets/ko-a001.jsonl` exists
    And it contains at least one JSON line

  Scenario: build_start event is emitted at the start of a build
    Given ticket ko-a001 is built
    Then the JSONL file contains an event with type "build_start"
    And the event has fields: ticket, ts

  Scenario: node_start and node_complete events bracket each node
    Given ticket ko-a001 is built through nodes triage and implement
    Then the JSONL file contains "node_start" events for triage and implement
    And the JSONL file contains "node_complete" events for triage and implement
    And each node_complete has fields: workflow, node, result

  Scenario: build_complete event records the terminal outcome
    Given ticket ko-a001 is built successfully
    Then the JSONL file contains a "build_complete" event
    And the event has field outcome = "succeed"

  Scenario: failed build records outcome
    Given ticket ko-a001 fails at a decision node
    Then the JSONL file contains a "build_complete" event
    And the event has field outcome = "fail"

  Scenario: JSONL persists across multiple builds
    Given ticket ko-a001 is built and fails
    And ticket ko-a001 is reset to open
    When I run `ko agent build ko-a001` again
    Then the JSONL file contains two "build_start" events
    And both builds' events are present

  Scenario: JSONL survives ticket close
    Given ticket ko-a001 is built successfully and closed
    Then `.ko/tickets/ko-a001.jsonl` still exists
    And the artifact directory has been removed

  Scenario: build history path is available as $KO_BUILD_HISTORY
    Given a pipeline with a run node that reads $KO_BUILD_HISTORY
    When the build runs
    Then the run node can access the build history file path

  Scenario: Retry attempts emit node_fail and node_retry events
    Given a pipeline with max_retries: 2
    And a node that fails on the first two attempts and succeeds on the third
    When I run `ko agent build ko-a001`
    Then the JSONL file contains 2 "node_fail" events with attempt numbers 1 and 2
    And the JSONL file contains 2 "node_retry" events with attempt numbers 2 and 3
    And the final "node_complete" event has result "done"

  Scenario: Exhausted retries emit node_fail events
    Given a pipeline with max_retries: 2
    And a node that always fails
    When I run `ko agent build ko-a001`
    Then the build fails
    And the JSONL file contains 3 "node_fail" events with attempt numbers 1, 2, and 3
    And the JSONL file contains 2 "node_retry" events (not after final failure)
    And the final "node_complete" event has result "error"

  Scenario: Hook failures emit build_error events
    Given a pipeline with an on_succeed hook that exits non-zero
    When I run `ko agent build ko-a001`
    Then the build fails and ticket is blocked
    And the JSONL file contains a "build_error" event with stage "on_succeed_hook"
    And the "build_complete" event has outcome "succeed" (workflow finished before hook failed)
