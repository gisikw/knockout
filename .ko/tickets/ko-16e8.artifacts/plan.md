## Goal
Add `ko answer` command to submit answers to plan questions and unblock tickets.

## Context

The codebase already has:
- **PlanQuestion schema** in `ticket.go:39-52` with `ID`, `Question`, `Context`, and `Options` fields
- **Question parsing and serialization** in `ticket.go:111-127` (FormatTicket) and `ticket.go:155-297` (ParseTicket)
- **ko block --questions** command in `cmd_status.go:105-181` that sets plan-questions and blocks tickets
- **AddNote helper** in `ticket.go:613-623` that appends timestamped notes to ticket body
- **Command structure pattern** follows `resolveProjectTicketsDir(args)` → `ResolveID()` → `LoadTicket()` → modify → `SaveTicket()` → `EmitMutationEvent()` (see `cmd_note.go`, `cmd_status.go`, etc.)
- **Main command router** in `main.go:23-83` uses simple string switch for subcommand dispatch

The ticket specifies:
- Command format: `ko answer <id> '<json>'` where JSON maps question IDs to answer strings
- Partial answers supported — only resolve questions whose IDs appear in the JSON
- Each answered question removed from `plan-questions` in frontmatter
- Each answer appended as timestamped note: `Plan answer (q1): Tabs or spaces? → Spaces, 2-wide`
- When last question answered, clear `plan-questions` entirely and set status to `open`
- If questions remain, status stays `blocked`

## Approach

Create a new `cmd_answer.go` file with `cmdAnswer()` following existing command patterns. The function will:
1. Parse args to extract ticket ID and JSON payload
2. Unmarshal JSON into a `map[string]string` (question ID → answer text)
3. Load the ticket and validate that answered question IDs exist
4. For each answered question, append a timestamped note with format `Plan answer (qID): Question text → Answer text`
5. Remove answered questions from `ticket.PlanQuestions` slice
6. If all questions answered, clear `plan-questions` and set status to `open`; otherwise keep status as `blocked`
7. Save ticket and emit mutation event

## Tasks

1. **[cmd_answer.go]** — Create new file with `cmdAnswer()` function.
   - Parse args: expect `<id>` and `<json>` (JSON string mapping question IDs to answers)
   - Use `resolveProjectTicketsDir()` to get tickets directory
   - Use `ResolveID()` to expand partial ticket ID
   - Use `json.Unmarshal()` to parse JSON into `map[string]string`
   - Validate that all question IDs in the JSON exist in `ticket.PlanQuestions`
   - For each answered question: build note text `Plan answer (qID): Question → Answer`, call `AddNote()`
   - Filter `ticket.PlanQuestions` to remove answered questions
   - If `ticket.PlanQuestions` becomes empty, set `ticket.Status = "open"` and clear the slice entirely
   - Call `SaveTicket()` and `EmitMutationEvent()` with event type "answer"
   - Print confirmation message with count of answered questions
   Verify: `go build` succeeds.

2. **[main.go:23-83]** — Add `"answer"` case to command switch, calling `cmdAnswer(rest)`.
   Verify: `go build` succeeds, help text updated if needed.

3. **[cmd_answer.go (tests)]** — Add `cmd_answer_test.go` with table-driven tests.
   - Test partial answer (some questions answered, others remain, status stays `blocked`)
   - Test full answer (all questions answered, `plan-questions` cleared, status set to `open`)
   - Test invalid JSON (should error)
   - Test nonexistent question ID (should error)
   - Test ticket with no plan-questions (should error)
   - Verify notes are appended correctly with proper format
   Verify: `go test -v -run TestCmdAnswer` passes.

## Open Questions

None. The ticket specifies exact behavior, JSON format, note format, and status transition logic. The codebase already has all necessary primitives (`AddNote`, `PlanQuestion` parsing/serialization, command structure patterns). Implementation is straightforward.
