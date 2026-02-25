## Goal
Remove 'create' command alias, keeping only 'add' for ticket creation.

## Context
Both 'ko add' and 'ko create' currently call the same `cmdCreate` function in main.go (lines 24-27). This is an intentional redundancy that the ticket wants to eliminate, keeping the shorter, punchier 'add' command.

Key files:
- **main.go**: Command dispatcher with switch statement routing both 'add' and 'create' to `cmdCreate`
- **cmd_create.go**: The actual implementation (cmdCreate function)
- **cmd_create_test.go**: Unit tests for ticket creation
- **specs/ticket_creation.feature**: Behavioral specs that reference both 'ko add' and 'ko create'
- **testdata/ticket_creation/*.txtar**: Integration tests using testscript format

The specs show both commands being used interchangeably (e.g., line 10 uses "ko create", line 28 uses "ko add"). Tests in testdata also use both forms.

## Approach
1. Remove the 'create' case from the command dispatcher in main.go
2. Update cmdHelp to remove 'create' from usage output
3. Update specs to use only 'ko add'
4. Update integration tests to use only 'ko add'
5. Keep cmd_create.go and cmdCreate function unchanged (implementation stays the same, just route changes)

No changes needed to cmd_create_test.go since it tests the function directly, not the command routing.

## Tasks
1. [main.go:26-27] — Remove the `case "create":` line and its `return cmdCreate(rest)`. Keep only the 'add' case.
   Verify: `go build` succeeds.

2. [main.go:cmdHelp] — Update help text to remove any mention of 'create' command. The command list currently shows both, should show only 'add'.
   Verify: `ko help` displays only 'add' in command list.

3. [specs/ticket_creation.feature] — Replace all instances of "ko create" with "ko add" in scenario steps. This maintains behavioral specs while using the canonical command.
   Verify: Spec scenarios still describe the same behavior, just with updated command syntax.

4. [testdata/ticket_creation/*.txtar] — Search for `exec ko create` and replace with `exec ko add` in all test files under testdata/ticket_creation/.
   Verify: `go test ./... -run TestCreate` passes (testscript tests).

5. [README.md] — Search for references to 'ko create' and replace with 'ko add'. Update any command examples or documentation.
   Verify: Documentation is consistent with implemented commands.

6. [Other testdata files] — Search all remaining testdata/*.txtar files for `ko create` usage and replace with `ko add`.
   Verify: `go test ./...` passes completely.

7. [Other spec files] — Search specs/*.feature files for "ko create" and replace with "ko add".
   Verify: All specs are consistent.

## Open Questions
None. This is a straightforward command alias removal with no ambiguity. The implementation function stays the same, only the routing changes. All tests and documentation need updating to reflect the single canonical command name.
