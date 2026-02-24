## Goal
Enable `ko add` to accept ticket description via stdin/heredoc or second positional arg.

## Context
The ticket depends on ko-90ba (rename `create` to `add`), which is already closed. Both `ko add` and `ko create` now map to `cmdCreate()` in cmd_create.go:12.

Current implementation in cmd_create.go:
- Line 40-43: Title is taken from fs.Arg(0)
- Line 90-92: The `-d` flag sets description, appending to t.Body
- No stdin handling
- No second positional arg handling

Reference implementation in cmd_note.go (lines 28-52):
- Checks if stdin is a pipe using `os.Stdin.Stat()` and `os.ModeCharDevice`
- Reads from stdin with `io.ReadAll(os.Stdin)` if piped
- Falls back to args[1:] if stdin is empty or stdin is a terminal
- Final validation ensures non-empty content

The project follows strict testing patterns per INVARIANTS.md:
- Specs in specs/*.feature (gherkin)
- Tests in testdata/*.txtar (testscript)
- Tests mirror source files (cmd_create.go → cmd_create_test.go)

## Approach
Add stdin detection and second positional arg handling to cmdCreate(), following the same pattern used in cmdAddNote(). The description sources (in priority order): stdin (if piped), second positional arg, `-d` flag. Merge all sources into t.Body, matching the current behavior where `-d` appends.

## Tasks
1. [cmd_create.go:cmdCreate] — After line 43 (title extraction), add stdin detection using the same pattern as cmd_note.go:30-52. Read from stdin if it's a pipe, otherwise check for fs.Arg(1) as the second positional arg. Store in a local `descFromArgsOrStdin` variable.
   Verify: `go build` compiles without errors.

2. [cmd_create.go:cmdCreate] — Around line 90-92 (where `-d` flag is checked), merge the description sources. If `descFromArgsOrStdin` is non-empty, append it to t.Body with newlines. Preserve the existing `-d` flag handling so both can be used together.
   Verify: Manual test with `ko add "test" "description"` creates a ticket with the description in the body.

3. [cmd_create_test.go] — Add test cases covering: (1) second positional arg sets description, (2) stdin (simulated pipe) sets description, (3) both `-d` flag and positional arg merge into body, (4) empty description is allowed (backward compat). Follow the pattern from TestReadWritePrefix and TestDetectPrefixPersists.
   Verify: `go test -run TestCreate` passes.

4. [specs/ticket_creation.feature] — If this spec exists, add scenarios for the new description input methods. If it doesn't exist, create it with basic scenarios including the new description handling.
   Verify: Spec is readable and matches implementation behavior.

## Open Questions
Should the description sources have a strict priority (stdin wins over arg, arg wins over `-d`) or should they all merge together? The current `-d` flag appends to t.Body, suggesting merge is the intended behavior. Assuming merge: stdin content + second arg + `-d` flag all append to t.Body in that order.

If multiple sources shouldn't merge, clarify the precedence order.
