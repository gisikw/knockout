Feature: Agent Harnesses
  As a system architect
  I want agent invocation to be handled by executable shell scripts
  So that different agent CLIs can be integrated regardless of their syntax

  # Shell Harness Architecture

  Scenario: Shell harnesses are executable scripts that receive KO_-namespaced environment variables
    Given a shell harness script at ".ko/agent-harnesses/test-agent"
    When the harness is invoked to run a prompt
    Then the script receives KO_PROMPT as an environment variable
    And the script receives KO_MODEL as an environment variable
    And the script receives KO_SYSTEM_PROMPT as an environment variable
    And the script receives KO_ALLOW_ALL as "true" or "false"
    And the script receives KO_ALLOWED_TOOLS as a comma-separated list

  Scenario: Shell harness scripts are marked executable
    Given a shell harness at ".ko/agent-harnesses/my-agent"
    Then the file has executable permissions
    And the file begins with a shebang (#!/bin/sh or similar)

  Scenario: Shell harness can handle binary fallback internally
    Given a shell harness that supports multiple binary names
    When the harness runs
    Then it uses "command -v" or similar to locate the binary
    And it falls back to alternative binary names if the first is not found

  Scenario: Built-in shell harnesses are embedded and extracted to temp location
    Given a built-in shell harness "claude"
    When the harness is loaded
    Then the embedded script is written to a temporary location
    And the temporary file has executable permissions (0755)
    And the temporary script is executed

  # Search Order and Precedence

  Scenario: Harness search order is project > user > built-in
    Given a built-in shell harness "claude"
    And a user shell harness at "~/.config/knockout/agent-harnesses/claude"
    And a project shell harness at ".ko/agent-harnesses/claude"
    When the "claude" harness is loaded
    Then the project harness is used (highest precedence)

  Scenario: Shell harnesses take precedence over YAML harnesses (during transition)
    Given a shell harness at ".ko/agent-harnesses/myagent"
    And a YAML harness at ".ko/agent-harnesses/myagent.yaml"
    When the "myagent" harness is loaded
    Then the shell harness is used
    And the YAML harness is ignored

  Scenario: Harness lookup checks for shell script before YAML file
    Given no file at ".ko/agent-harnesses/custom"
    And a YAML file at ".ko/agent-harnesses/custom.yaml"
    When the "custom" harness is loaded
    Then the YAML harness is used (backward compatibility)

  # Environment Variable Contract

  Scenario: KO_PROMPT contains the full prompt text
    Given a shell harness receives a prompt "Analyze this ticket"
    When the harness runs
    Then KO_PROMPT equals "Analyze this ticket"

  Scenario: KO_MODEL is set when a model is specified
    Given a pipeline with "model: opus"
    When the harness runs
    Then KO_MODEL equals "opus"

  Scenario: KO_MODEL is empty when no model is specified
    Given a pipeline with no model specified
    When the harness runs
    Then KO_MODEL is set to an empty string

  Scenario: KO_SYSTEM_PROMPT contains system prompt content
    Given a system prompt "You are a code reviewer"
    When the harness runs
    Then KO_SYSTEM_PROMPT equals "You are a code reviewer"

  Scenario: KO_ALLOW_ALL is "true" when all tools are allowed
    Given a pipeline with allow_all: true
    When the harness runs
    Then KO_ALLOW_ALL equals "true"

  Scenario: KO_ALLOW_ALL is "false" when specific tools are allowed
    Given a pipeline with allowed_tools: ["read", "write"]
    When the harness runs
    Then KO_ALLOW_ALL equals "false"

  Scenario: KO_ALLOWED_TOOLS is a comma-separated list of tool names
    Given a pipeline with allowed_tools: ["read", "write", "bash"]
    When the harness runs
    Then KO_ALLOWED_TOOLS equals "read,write,bash"

  Scenario: KO_ALLOWED_TOOLS is empty when allow_all is true
    Given a pipeline with allow_all: true
    When the harness runs
    Then KO_ALLOWED_TOOLS is set to an empty string

  # Prompt Passing

  Scenario: Shell harness decides how to pass prompt to the agent
    Given a shell harness receives KO_PROMPT
    When the harness constructs its agent command
    Then the harness may pass the prompt via stdin
    Or the harness may pass the prompt as a command argument
    Or the harness may pass the prompt using agent-specific flags
    # The harness has full control over how to invoke the agent

  Scenario: Built-in claude harness passes prompt via stdin
    Given the built-in "claude" harness
    And KO_PROMPT is "Fix this bug"
    When the harness runs
    Then the prompt is echoed to stdin of the claude command

  Scenario: Built-in cursor harness passes prompt as -p argument
    Given the built-in "cursor" harness
    And KO_PROMPT is "Implement feature"
    When the harness runs
    Then the prompt is passed as "-p 'Implement feature'"

  # Migration from YAML

  Scenario: YAML harness templates are deprecated
    Given the system is using shell harnesses
    When a user creates a new custom harness
    Then they must create a shell script
    And YAML harness format is no longer supported

  Scenario: Existing YAML harnesses must be migrated to shell scripts
    Given a user has a custom YAML harness
    When they upgrade to the shell harness version
    Then they must convert the YAML to a shell script
    And the shell script receives parameters via KO_* environment variables
    And the shell script constructs the agent command invocation
