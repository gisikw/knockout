## Goal
Enable `ko add` to accept ticket description via stdin (heredoc) or second positional arg with strict priority: stdin > second arg > -d flag.

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
- Falls back to args if stdin is empty or stdin is a terminal
- Final validation ensures non-empty content (for notes — not required for descriptions)

**Decision clarified (2026-02-24 13:02:31 UTC):** Strict priority, no merging. stdin > second positional arg > -d flag. Only the highest-priority non-empty source is used.

The project follows strict testing patterns per INVARIANTS.md:
- Specs in specs/*.feature (gherkin)
- Tests in testdata/*.txtar (testscript)
- Tests mirror source files (cmd_create.go → cmd_create_test.go)

## Approach
Add stdin detection and second positional arg handling to cmdCreate() following the pattern in cmdAddNote(). Implement strict priority: if stdin has content, use only stdin; else if second positional arg exists, use only that; else use `-d` flag if present. The `-d` flag becomes the lowest-priority fallback. Empty descriptions are valid (backward compatibility).

## Tasks
1. [cmd_create.go:cmdCreate] — After title extraction (line 43), add description source detection. Check stdin first using the same pattern as cmd_note.go:30-40. If stdin is a pipe and non-empty, use it. Otherwise check fs.Arg(1) for second positional arg. Store the result in a local `descFromInput` variable. Add `io` import if needed.
   Verify: `go build` compiles without errors.

2. [cmd_create.go:cmdCreate] — Refactor the description handling block (lines 90-92). Apply strict priority: if `descFromInput` is non-empty, use it and ignore the `-d` flag. Otherwise fall back to `-d` flag. Preserve the newline wrapping behavior: `t.Body += "\n" + description + "\n"`.
   Verify: Manual test with `ko add "title" "description"` creates a ticket with description. Test that `-d` is ignored when second arg is present.

3. [cmd_create_test.go] — Add test function `TestCreateWithDescription` covering: (1) second positional arg sets description, (2) `-d` flag sets description when no positional arg, (3) second arg takes priority over `-d` flag (flag is ignored), (4) empty description is allowed (no args, no flag). Follow the pattern from TestCreateWithShorthandPriority. Use a temp directory, create tickets, verify description content in ticket body.
   Verify: `go test -run TestCreateWithDescription` passes.

4. [cmd_create_test.go] — Add test function `TestCreateWithStdinDescription` to test stdin handling. Mock stdin by temporarily redirecting it to a pipe with known content. Verify stdin sets description and takes priority over both positional arg and `-d` flag. This is more complex — may need to use `os.Pipe()` or skip if too brittle.
   Verify: `go test -run TestCreateWithStdin` passes or is appropriately skipped with a comment explaining why stdin mocking is deferred.

5. [specs/ticket_creation.feature] — Check if this spec exists. If it does, add scenarios for: (1) creating ticket with second positional arg description, (2) creating ticket with stdin description, (3) priority behavior (stdin > arg > flag). If the spec doesn't exist, create it with these scenarios plus basic creation scenario. Follow gherkin syntax (Given/When/Then).
   Verify: Spec is readable, follows gherkin syntax, and accurately describes the new behavior.

## Open Questions
None. The priority decision has been clarified: strict priority with stdin > second positional arg > -d flag. Empty descriptions remain valid for backward compatibility.
