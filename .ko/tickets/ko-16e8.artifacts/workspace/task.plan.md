The plan is complete and written to the artifact directory. Let me provide a summary:

## Plan Summary

I've written an implementation plan for adding the `ko answer` command to submit answers to plan questions. The plan includes:

**Goal**: Add `ko answer <id> '<json>'` command to unblock tickets by answering plan questions.

**Key findings from codebase investigation**:
- PlanQuestion schema already exists in `ticket.go` with full parsing/serialization support
- `ko block --questions` already implemented in `cmd_status.go`
- Clear command patterns to follow (seen in `cmd_note.go`, `cmd_status.go`)
- `AddNote()` helper available for timestamped note appending

**Implementation approach**:
1. Create `cmd_answer.go` with standard command structure
2. Parse JSON mapping question IDs to answer strings
3. Append timestamped notes for each answer: `Plan answer (q1): Question → Answer`
4. Remove answered questions from `plan-questions` frontmatter
5. When all questions answered, clear `plan-questions` and set status to `open`
6. When questions remain, keep status as `blocked`

**Tasks breakdown**:
1. Create `cmd_answer.go` with validation and note-appending logic
2. Wire into `main.go` command switch
3. Add comprehensive tests in `cmd_answer_test.go`

**Open questions**: None — the ticket specifies exact behavior and all necessary primitives exist in the codebase.
