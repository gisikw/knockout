## Goal
Replace `ko block`, `ko questions`, and `ko answer` commands with a single `ko triage` command that handles blocking, questions, and answers.

## Context
The existing implementation has three separate commands:
- `cmd_status.go:cmdBlock` - Sets status to blocked, optionally accepts `--questions` flag with JSON
- `cmd_questions.go:cmdQuestions` - Displays plan questions as JSON
- `cmd_answer.go:cmdAnswer` - Accepts answers JSON, adds notes, removes answered questions, unblocks when all answered

Key patterns found:
- Commands use `resolveProjectTicketsDir(args)` to find tickets directory and support #project tags
- Status transitions emit mutation events via `EmitMutationEvent`
- `PlanQuestions` field on `Ticket` struct holds array of questions (YAML in frontmatter)
- `cmdAnswer` validates questions exist, creates notes with timestamp, and sets status to "open" when all questions answered
- `ExtractBlockReason` parses notes formatted as "ko: BLOCKED — {reason}" to extract block reasons
- Tests follow pattern in `cmd_questions_test.go` and `cmd_answer_test.go` using temp directories

The revised design (from ticket notes) specifies:
- Bare `ko triage <id>` shows block reason + open questions
- `ko triage <id> --block [reason]` blocks with optional reason as note
- `ko triage <id> --questions 'json'` adds questions, implicitly sets status=blocked
- `ko triage <id> --answers 'json'` answers questions, unblocks when all answered

**Resolved design questions:**
1. Output format: Human-readable text for block reason, followed by JSON array for questions
2. Note format: Use "ko: BLOCKED — {reason}" format for `ExtractBlockReason` compatibility
3. Breaking change: Remove old commands entirely (no backward compatibility)

## Approach
Create a new `cmd_triage.go` file that consolidates the three commands. The command will use flag parsing to determine which mode to operate in (show/block/questions/answers). When no flags are provided, it displays current triage state using `ExtractBlockReason` for the block reason and JSON output for questions. The implementation will reuse existing logic from `cmdBlock`, `cmdQuestions`, and `cmdAnswer`, maintaining the same data model and mutation event patterns. Old commands and their tests will be removed entirely.

## Tasks
1. [cmd_triage.go] — Create new file with `cmdTriage` function that:
   - Parses flags: `--block` (optional string), `--questions` (string), `--answers` (string)
   - Bare invocation: calls `ExtractBlockReason` to show block reason as text, then outputs `PlanQuestions` as JSON array (reuse pattern from `cmdQuestions`)
   - `--block`: sets status to blocked, adds note "ko: BLOCKED — {reason}" if reason provided (matching `ExtractBlockReason` format)
   - `--questions`: validates JSON with `ValidatePlanQuestions`, sets `PlanQuestions`, sets status to blocked, emits mutation event
   - `--answers`: reuses validation and note-adding logic from `cmdAnswer` lines 28-84, removes answered questions, auto-unblocks when all answered
   - Uses `resolveProjectTicketsDir` and emits appropriate mutation events
   Verify: `go build` succeeds.

2. [main.go:run] — Add `case "triage":` that calls `cmdTriage(rest)` (around line 38).
   Verify: `go build` succeeds.

3. [cmd_triage_test.go] — Create comprehensive tests covering:
   - Bare invocation showing block reason and questions as JSON
   - `--block` with reason (verify note format)
   - `--block` without reason (bare block)
   - `--questions` setting questions and auto-blocking
   - `--answers` partial (some questions remain) and full (auto-unblock)
   - Error cases: invalid JSON, missing ticket, question ID not found, no questions to answer
   Pattern tests after `cmd_questions_test.go` and `cmd_answer_test.go` using temp directories.
   Verify: `go test ./... -run TestCmdTriage` passes.

4. [main.go:run] — Remove switch cases for `"block"` (line 38), `"questions"` (line 56), `"answer"` (line 52).
   Verify: `go build` succeeds.

5. [cmd_status.go] — Remove `cmdBlock` function (lines 105-181).
   Verify: `go build` succeeds (no references to removed function).

6. [cmd_questions.go] — Delete entire file.
   Verify: `go build` succeeds.

7. [cmd_questions_test.go] — Delete entire file.
   Verify: `go test ./...` still passes (functionality moved to cmd_triage_test.go).

8. [cmd_answer.go] — Delete entire file.
   Verify: `go build` succeeds.

9. [cmd_answer_test.go] — Delete entire file.
   Verify: `go test ./...` still passes (functionality moved to cmd_triage_test.go).

10. [main.go:cmdHelp] — Update help text to replace old commands with `ko triage`:
    - Remove lines for: `block <id>`, `block <id> --questions '<json>'`, `questions <id>`, `answer <id> '<json>'`
    - Add: `triage <id>` - show block reason and open questions
    - Add: `triage <id> --block [reason]` - block ticket with optional reason
    - Add: `triage <id> --questions '<json>'` - add questions, implicitly blocks
    - Add: `triage <id> --answers '<json>'` - answer questions, auto-unblock when done
    Verify: `go build && ./ko help` shows updated help text.

11. [README.md] — Update command list to replace old commands with `ko triage` (if documented).
    Verify: Markdown renders correctly, no broken references to old commands.

## Open Questions
None. All design questions have been answered by the ticket author:
- Output format: human-readable + JSON (answered)
- Note format: "ko: BLOCKED — {reason}" for parseability (answered)
- Breaking change: yes, remove old commands entirely (answered)
