## Goal
Add `--json` flag support to all commands that produce human-readable output.

## Context
The codebase has several commands that produce human-readable output. Some already have `--json` support:
- `show` (cmd_show.go) - has `--json` flag returning showJSON struct
- `ls` (cmd_list.go) - has `--json` flag returning ticketJSON array
- `ready` (cmd_list.go) - has `--json` flag returning ticketJSON array
- `blocked` (cmd_list.go) - has `--json` flag returning ticketJSON array

Commands that need `--json` support:
- `agent status` (cmd_agent.go:248-285) - currently outputs human-readable status text
- `triage` (cmd_triage.go:99-122) - shows block reason and questions, questions already JSON but reason is text
- `dep tree` (cmd_dep.go:136-175) - outputs indented tree structure
- `project ls` (cmd_project.go:137-172) - outputs project list with markers

Commands that don't need `--json` (write-only operations that output simple confirmations):
- `add`, `close`, `start`, `open`, `status`, `dep`, `undep`, `note`, `bump`, `clear`, `update`, `block`, `project set`, `agent build`, `agent loop`, `agent init`, `agent start`, `agent stop`, `serve`

Testing pattern: The codebase uses standard Go table-driven tests. Tests are in corresponding `*_test.go` files. Existing JSON support tests redirect stdout and check the output (see cmd_status_test.go).

INVARIANTS.md specifies:
- CLI errors go to stderr with non-zero exit code
- Every new behavior gets a spec and test
- Decision logic should be pure where possible
- 500 line max per file (cmd_agent.go is already at 304 lines, cmd_list.go at 428 lines)

## Approach
For each command that needs JSON support, add a `--json` flag using the existing pattern (flag.FlagSet with flag.Bool). When `--json` is true, output structured JSON to stdout instead of human-readable text. For `agent status`, define a new struct type. For `triage`, wrap the existing output in a struct. For `dep tree` and `project ls`, define appropriate struct types for the tree/list data.

## Tasks
1. [cmd_agent.go:cmdAgentStatus] — Add `--json` flag to `agent status` command. Define `agentStatusJSON` struct with fields for provisioned status, running status, pid, and last log line. When `--json` is true, output JSON; otherwise keep existing text format.
   Verify: `go test ./... -v -run TestCmdAgentStatus` passes (after adding test).

2. [cmd_agent_test.go] — Add table-driven test `TestCmdAgentStatusJSON` that verifies JSON output format includes correct fields and values for various agent states (not provisioned, not running, running with pid).
   Verify: `go test ./... -v -run TestCmdAgentStatusJSON` passes.

3. [cmd_triage.go:showTriageState] — Add `--json` flag parameter to `showTriageState` function and `cmdTriage`. Define `triageStateJSON` struct with fields for block_reason and questions. When `--json` is true, output single JSON object; otherwise keep existing text format.
   Verify: `go test ./... -v -run TestCmdTriage` passes (update existing test if needed).

4. [cmd_triage_test.go] — Add test cases to existing `TestCmdTriage` or create `TestCmdTriageJSON` to verify JSON output format for triage show mode with and without block reason/questions.
   Verify: `go test ./... -v -run TestCmdTriage` passes.

5. [cmd_dep.go:cmdDepTree] — Add `--json` flag to `dep tree` command. Define `depTreeJSON` struct with fields for id, status, title, and deps (recursive). Add pure function `buildDepTree(ticketsDir, id, visited) *depTreeJSON` that constructs the tree structure. When `--json` is true, call buildDepTree and encode to JSON; otherwise keep existing printDepTree behavior.
   Verify: `go test ./... -v -run TestCmdDepTree` passes (after adding test).

6. [cmd_dep_test.go] — Create new file with `TestCmdDepTreeJSON` that verifies JSON output structure for various dependency tree shapes (single node, linear chain, branching tree, cycle detection).
   Verify: `go test ./... -v -run TestCmdDepTreeJSON` passes.

7. [cmd_project.go:cmdProjectLs] — Add `--json` flag to `project ls` command. Define `projectJSON` struct with fields for tag, path, and is_default bool. When `--json` is true, output array of projectJSON; otherwise keep existing tabular format.
   Verify: `go test ./... -v -run TestCmdProjectLs` passes (update existing test).

8. [cmd_project_test.go] — Add test cases to existing project tests or create `TestCmdProjectLsJSON` to verify JSON output includes all registered projects with correct default marker.
   Verify: `go test ./... -v -run TestCmdProjectLs` passes.

9. [README.md] — Update command documentation to note `--json` flag availability for `show`, `ls`, `ready`, `blocked`, `triage`, `agent status`, `dep tree`, and `project ls` commands.
   Verify: Manual review that documentation is accurate and complete.

## Open Questions
None. The pattern is established and the requirements are clear from the ticket.
