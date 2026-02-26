## Goal
Remove the `ko clear` command entirely from the codebase.

## Context
The `ko clear --force` command is a destructive operation that removes all local tickets from `.ko/tickets/`. It's implemented in `cmd_clear.go` and registered in `main.go`'s command dispatcher (line 58-59). The command also appears in:
- The help text in `main.go` (line 143)
- The README.md documentation (line 39)

No tests or specs exist for this command, so removal is clean with no test changes needed.

The ticket notes that archival/cleanup functionality deserves more careful design before exposing a 'delete everything' command, making this a pure removal task.

## Approach
Remove the command file, unregister it from the main dispatcher, and remove it from all documentation. This is a straightforward deletion with no replacement functionality.

## Tasks
1. [cmd_clear.go] — Delete the entire file.
   Verify: File no longer exists in the repository.

2. [main.go:58-59] — Remove the `case "clear":` block from the command dispatcher.
   Verify: `go build` succeeds.

3. [main.go:143] — Remove the `clear --force` line from the help text.
   Verify: `ko help` output no longer mentions clear command.

4. [README.md:39] — Remove the `clear --force` line from the README usage section.
   Verify: README renders correctly and has no references to clear.

5. Verify: Run `go test ./...` to ensure no tests break.
   Verify: All tests pass.

## Open Questions
None. This is a clean removal with no dependencies or replacement functionality needed.
