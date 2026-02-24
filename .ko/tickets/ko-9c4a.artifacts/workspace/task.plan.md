The plan is complete and written to `/home/dev/Projects/knockout/.ko/tickets/ko-9c4a.artifacts/plan.md`. 

The plan breaks down the implementation into 7 concrete tasks:

1. **Add struct fields** — Extend `Ticket` with `PlanQuestions` and related types
2. **Serialization** — Update `FormatTicket()` to write questions as YAML
3. **Parsing** — Update `ParseTicket()` to read questions from frontmatter using state machine
4. **Tests** — Add table-driven tests for round-trip parsing/formatting
5. **CLI flag** — Implement `--questions` flag in `cmdBlock()` with JSON parsing
6. **Validation** — Add pure validation function for question structure
7. **Spec** — Add gherkin scenarios documenting the behavior

The approach uses the existing minimal YAML parser pattern (no external dependencies) and follows the codebase's invariants: pure decision logic, fail-fast on bad input, and specs before code. Plain `ko block <id>` will continue to work without questions.
