## Summary

Two one-line changes in `cmd_agent.go`:

- **Line 178** (`cmdAgentStart`): `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for "agent started (pid %d), logging to %s"
- **Line 227** (`cmdAgentStop`): `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for "agent stopped (pid %d)"

Both status messages now go to stderr, keeping stdout clean for composable use. This follows standard CLI convention (status/diagnostic messages on stderr, data on stdout).

**Notable decisions:**
- `ko agent stop` was included in this ticket (not deferred) for consistency, as confirmed by the author.
- The `maybeAutoAgent` function referenced in the first-pass plan does not exist in the codebase; the prior plan's Task 2 about updating its comment was correctly dropped.

**Build and tests:** `go build ./...` and `go test ./...` both pass with no issues.
