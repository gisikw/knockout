Feature: HTTP Server (ko serve)

  Background:
    The `ko serve` command starts an HTTP daemon that accepts ko subcommand
    execution requests via a POST /ko endpoint. This enables programmatic
    access to ko functionality without spawning new processes for each command.

  Rule: Server lifecycle

    Scenario: Start server on default port
      Given a knockout project exists
      When I run "ko serve"
      Then the server starts on port 9876
      And stdout shows "ko serve: listening on :9876"

    Scenario: Start server on custom port
      Given a knockout project exists
      When I run "ko serve -p 8080"
      Then the server starts on port 8080
      And stdout shows "ko serve: listening on :8080"

    Scenario: Graceful shutdown on SIGTERM
      Given a knockout project exists
      And a ko serve instance is running
      When I send SIGTERM to the process
      Then the server shuts down gracefully within 5 seconds
      And stdout shows "received interrupt, shutting down"
      And the exit code is 0

    Scenario: Graceful shutdown on SIGINT
      Given a knockout project exists
      And a ko serve instance is running
      When I send SIGINT to the process
      Then the server shuts down gracefully within 5 seconds
      And stdout shows "received interrupt, shutting down"
      And the exit code is 0

  Rule: Endpoint validation

    Scenario: Only POST method allowed
      Given a ko serve instance is running
      When I send a GET request to "/ko"
      Then the response status is 405 Method Not Allowed

    Scenario: Only /ko endpoint exists
      Given a ko serve instance is running
      When I send a POST request to "/other"
      Then the response status is 404 Not Found

  Rule: Request body validation

    Scenario: Valid JSON with argv array
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["ready"]}
        """
      Then the response status is 200
      And the response Content-Type is "text/plain"

    Scenario: Invalid JSON
      Given a ko serve instance is running
      When I POST to "/ko" with body "not json"
      Then the response status is 400
      And the response body contains "invalid JSON"

    Scenario: Empty argv array
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": []}
        """
      Then the response status is 400
      And the response body contains "argv must have at least one element"

    Scenario: Missing argv field
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"command": "ready"}
        """
      Then the response status is 400

  Rule: Subcommand whitelist enforcement

    Scenario Outline: Whitelisted subcommands are allowed
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["<subcommand>"]}
        """
      Then the response status is 200 or 400
      # 200 if command succeeds, 400 if command fails
      # (depends on project state, not whitelist rejection)

      Examples:
        | subcommand  |
        | ls          |
        | ready       |
        | blocked     |
        | resolved    |
        | closed      |
        | query       |
        | show        |
        | questions   |
        | answer      |
        | close       |
        | open        |
        | block       |
        | start       |
        | bump        |
        | note        |
        | status      |
        | dep         |
        | undep       |
        | agent       |

    Scenario Outline: Non-whitelisted subcommands are rejected
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["<subcommand>"]}
        """
      Then the response status is 400
      And the response Content-Type is "application/json"
      And the response body contains:
        """
        {"error": "subcommand '<subcommand>' not allowed"}
        """

      Examples:
        | subcommand |
        | rm         |
        | eval       |
        | exec       |
        | create     |
        | add        |
        | init       |
        | version    |

  Rule: Command execution

    Scenario: Successful command execution
      Given a knockout project exists with tickets:
        | id      | status |
        | ko-1234 | open   |
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["ready"]}
        """
      Then the response status is 200
      And the response Content-Type is "text/plain"
      And the response body contains ticket output

    Scenario: Failed command execution
      Given a knockout project exists
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["show", "nonexistent-id"]}
        """
      Then the response status is 400
      And the response Content-Type is "application/json"
      And the response body contains:
        """
        {"error": "..."}
        """

    Scenario: Command with multiple arguments
      Given a knockout project exists with ticket ko-1234
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["show", "ko-1234"]}
        """
      Then the response status is 200
      And the response body contains ticket ko-1234 details

  Rule: Project-scoped execution

    Scenario: Request with #tag resolves to registered project
      Given a registry with projects:
        | tag      | path              |
        | knockout | /home/user/knockout |
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["ls"], "project": "#knockout"}
        """
      Then the response status is 200
      And the ko ls command executes in /home/user/knockout

    Scenario: Request with absolute path uses it directly
      Given a knockout project exists at /home/user/myproject
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["ready"], "project": "/home/user/myproject"}
        """
      Then the response status is 200
      And the ko ready command executes in /home/user/myproject

    Scenario: Request with invalid tag returns error
      Given a registry with no project "unknown"
      And a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["ls"], "project": "#unknown"}
        """
      Then the response status is 404
      And the response Content-Type is "application/json"
      And the response body contains:
        """
        {"error": "project not found: #unknown"}
        """

    Scenario: Request without project uses cwd
      Given a knockout project exists at /path/to/project
      And I start ko serve from /path/to/project
      When I POST to "/ko" with body:
        """
        {"argv": ["ready"]}
        """
      Then the response status is 200
      And the ko ready command executes in /path/to/project

    Scenario: Request with empty project string uses cwd
      Given a knockout project exists at /path/to/project
      And I start ko serve from /path/to/project
      When I POST to "/ko" with body:
        """
        {"argv": ["ready"], "project": ""}
        """
      Then the response status is 200
      And the ko ready command executes in /path/to/project

  Rule: Security constraints

    Scenario: Cannot execute arbitrary shell commands
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["sh", "-c", "rm -rf /"]}
        """
      Then the response status is 400
      And the response body contains "subcommand 'sh' not allowed"
      And no shell commands are executed

    Scenario: Cannot execute system utilities
      Given a ko serve instance is running
      When I POST to "/ko" with body:
        """
        {"argv": ["cat", "/etc/passwd"]}
        """
      Then the response status is 400
      And the response body contains "subcommand 'cat' not allowed"

  Rule: Process isolation

    Scenario: Each request spawns separate ko process
      Given a ko serve instance is running
      When I POST to "/ko" with body '{"argv": ["ready"]}'
      Then a new "ko ready" process is spawned
      And the process uses the same binary as the server (os.Args[0])
      And the process inherits the server's working directory

    Scenario: Working directory is preserved
      Given a knockout project exists at /path/to/project
      And I start ko serve from /path/to/project
      When I POST to "/ko" with body '{"argv": ["ls"]}'
      Then the ko ls command executes in /path/to/project
