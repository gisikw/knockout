Done. Two one-line changes:

- **Task 1** (`cmd_agent.go:178`): `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for "agent started" message
- **Task 2** (`cmd_agent.go:227`): `fmt.Printf` → `fmt.Fprintf(os.Stderr, ...)` for "agent stopped" message

Build passes. No deviations from the plan.
