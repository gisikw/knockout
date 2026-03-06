Plan updated. The fix is two one-line changes:

- **`cmd_agent.go:178`**: `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for the "agent started" message
- **`cmd_agent.go:227`**: `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for the "agent stopped" message

Both open questions from the prior plan are resolved (include stop in this ticket, stderr for direct invocations is fine). The `maybeAutoAgent` function referenced in the earlier plan doesn't exist in the codebase, so that task is dropped. No new tests needed.
