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
- Tests follow pattern in `cmd_questions_test.go` and `cmd_answer_test.go` using temp directories

The revised design (from ticket notes) specifies:
- Bare `ko triage <id>` shows block reason + open questions
- `ko triage <id> --block [reason]` blocks with optional reason as note
- `ko triage <id> --questions 'json'` adds questions, implicitly sets status=blocked
- `ko triage <id> --answers 'json'` answers questions, unblocks when all answered

## Approach
Create a new `cmd_triage.go` file that consolidates the three commands. The command will use flag parsing to determine which mode to operate in (show/block/questions/answers). When no flags are provided, it displays current triage state. The implementation will reuse existing logic from `cmdBlock`, `cmdQuestions`, and `cmdAnswer`, maintaining the same data model and mutation event patterns.

## Tasks
1. [cmd_triage.go] — Create new file with `cmdTriage` function that:
   - Parses flags: `--block` (optional string), `--questions` (string), `--answers` (string)
   - Bare invocation: displays block reason (via `ExtractBlockReason`) and plan questions as JSON
   - `--block`: sets status to blocked, adds reason as note if provided
   - `--questions`: validates and sets `PlanQuestions`, sets status to blocked
   - `--answers`: reuses logic from `cmdAnswer` (validate, add notes, remove answered, unblock when done)
   - Uses `resolveProjectTicketsDir` and emits mutation events
   Verify: `go build` succeeds.

2. [main.go:run] — Add `case "triage":` that calls `cmdTriage(rest)`.
   Verify: `go build` succeeds.

3. [cmd_triage_test.go] — Create comprehensive tests covering:
   - Bare invocation showing block reason and questions
   - `--block` with and without reason
   - `--questions` setting questions and blocking
   - `--answers` partial and full (with auto-unblock)
   - Error cases (invalid JSON, missing ticket, etc.)
   Pattern tests after `cmd_questions_test.go` and `cmd_answer_test.go`.
   Verify: `go test ./... -run TestCmdTriage` passes.

4. [main.go:cmdHelp] — Update help text to replace:
   - Remove: `block <id>`, `questions <id>`, `answer <id> <json>` lines
   - Add: `triage <id>` entry explaining the new consolidated command
   Verify: `go build && ./ko help` shows updated help.

5. [README.md] — Update command list to replace old commands with `ko triage`.
   Verify: Markdown renders correctly.

6. [cmd_status.go] — Remove `cmdBlock` function (lines 105-181).
   Verify: `go build` succeeds (no references to removed function).

7. [cmd_questions.go] — Remove entire file.
   Verify: `go build` succeeds.

8. [cmd_questions_test.go] — Remove entire file.
   Verify: `go test ./...` passes (tests moved to cmd_triage_test.go).

9. [cmd_answer.go] — Remove entire file.
   Verify: `go build` succeeds.

10. [cmd_answer_test.go] — Remove entire file.
    Verify: `go test ./...` passes (tests moved to cmd_triage_test.go).

11. [main.go:run] — Remove switch cases for `"block"`, `"questions"`, `"answer"`.
    Verify: `go build` succeeds, commands return "unknown command" error.

## Open Questions
1. Should bare `ko triage <id>` output be human-readable or JSON-only? The ticket says "show block reason + open questions" but doesn't specify format. Current assumption: human-readable text showing block reason, followed by JSON array of questions (matching the pattern of existing `ko questions` output).

2. When using `--block` with a reason, should the note format match the existing "ko: BLOCKED" format that `ExtractBlockReason` parses? Current assumption: yes, use format "ko: BLOCKED — {reason}" so it's parseable by existing `ExtractBlockReason` function.

3. Should we maintain backward compatibility by keeping the old commands as aliases? The ticket says "Replaces" which implies removal, but existing users may have scripts. Current assumption: remove old commands entirely per ticket spec, but this is a breaking change worth noting.
