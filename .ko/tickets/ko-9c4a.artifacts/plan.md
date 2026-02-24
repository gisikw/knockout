## Goal
Add `plan-questions` field to ticket frontmatter schema and implement `--questions` flag for `ko block` command to write structured questions into frontmatter.

## Context
The codebase uses a minimal YAML parser for ticket frontmatter (no external dependencies). Tickets are defined in `ticket.go` with the `Ticket` struct and parsed via `ParseTicket()` / serialized via `FormatTicket()`.

Key files:
- `ticket.go:18-36` — `Ticket` struct definition with yaml tags for frontmatter fields
- `ticket.go:73-101` — `FormatTicket()` serializes Ticket to markdown with frontmatter
- `ticket.go:104-161` — `ParseTicket()` parses frontmatter using `parseYAMLLine()` and `parseYAMLList()`
- `cmd_status.go:79-85` — `cmdBlock()` currently just delegates to `cmdStatus()` with "blocked" status

The parser handles simple scalars (string, int) and string arrays (`deps`, `tags`) via inline bracket notation `[a, b, c]`. Complex nested structures (like the pipeline's node definitions) use indentation-based parsing with state machines.

Testing pattern: Go table-driven tests (see `ticket_test.go:5-64` for `TestIsReady`) and gherkin specs in `specs/*.feature`.

INVARIANTS.md requirements:
- Every behavior needs a spec in `specs/*.feature`
- File size limit: 500 lines max
- Decision logic must be pure (no I/O)
- Fail fast on bad input

## Approach
Extend the `Ticket` struct with a `PlanQuestions` field (array of question objects). Implement parsing and serialization for nested YAML using indentation-based state machine (similar to pipeline parsing). Add `--questions` flag to `cmdBlock()` that parses JSON input, validates the structure, writes it to the ticket frontmatter, and sets status to blocked.

The questions JSON will be parsed from the command line, then serialized to YAML in the frontmatter. Plain `ko block <id>` without `--questions` will continue to work as before (just sets status to blocked).

## Tasks
1. [ticket.go:Ticket] — Add `PlanQuestions` field to the Ticket struct as `[]PlanQuestion` with yaml tag `plan-questions,omitempty`. Define `PlanQuestion` and `QuestionOption` structs matching the schema (id, question, context, options array).
   Verify: `go build` succeeds, no type errors.

2. [ticket.go:FormatTicket] — Extend FormatTicket to serialize `PlanQuestions` to YAML. If the field is non-empty, write it as indented YAML with proper structure (2-space indent per level, arrays use `- ` prefix).
   Verify: Write a test ticket with plan-questions, verify it formats correctly with proper indentation.

3. [ticket.go:ParseTicket] — Extend ParseTicket to parse `plan-questions:` from frontmatter. Use indentation-based state machine similar to pipeline.go parsing. Track parse state (in_questions, in_options, current_question, current_option) and accumulate into the struct.
   Verify: Parse a ticket with plan-questions YAML, verify the struct matches expected values.

4. [ticket_test.go] — Add table-driven tests for ParseTicket and FormatTicket with plan-questions. Test cases: empty questions array, single question with 2 options, multiple questions, questions with context field, round-trip (parse → format → parse produces identical struct).
   Verify: `go test ./... -run TestPlanQuestions` passes.

5. [cmd_status.go:cmdBlock] — Replace the simple delegation to cmdStatus. Add flag parsing for `--questions` (JSON string). Parse the JSON into []PlanQuestion, validate required fields (id, question, options with label/value). If valid, load ticket, set PlanQuestions field, set status to blocked, save. If --questions is omitted, behave as before (just set status).
   Verify: `ko block <id> --questions '<json>'` writes questions to frontmatter and sets status. `ko block <id>` still works without questions.

6. [cmd_status.go or new file] — Add a pure validation function `ValidatePlanQuestions(questions []PlanQuestion) error` that checks required fields and returns helpful error messages for missing/invalid data.
   Verify: Unit test with valid and invalid question structures, verify error messages are clear.

7. [specs/ticket_creation.feature or new file] — Add gherkin scenarios for blocking with questions: "ko block without questions", "ko block with valid questions JSON", "ko block with invalid questions JSON fails".
   Verify: Spec describes expected behavior clearly, no implementation details.

## Open Questions
None — the schema is fully specified in the ticket, and the JSON-to-YAML conversion pattern is established in the codebase (see pipeline.go for reference).
