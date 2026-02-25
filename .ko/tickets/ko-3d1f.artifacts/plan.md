## Goal
Add a `ko questions` command that outputs plan-questions from ticket frontmatter as JSON.

## Context
The plan-questions schema was implemented in ko-9c4a. It's stored in the `PlanQuestions` field of the `Ticket` struct (ticket.go:29) as an array of `PlanQuestion` objects. Each question has:
- `ID` (required)
- `Question` (required)
- `Context` (optional)
- `Options` (array of `{Label, Value, Description?}`)

The `ko answer` command (ko-16e8) consumes these questions. The new `ko questions` command completes the read/write/update trio.

The codebase follows these patterns:
- Commands are implemented as `cmdXxx(args []string) int` functions in `cmd_*.go` files
- Commands use `resolveProjectTicketsDir(args)` to handle cross-project routing
- `ResolveID()` handles partial ticket ID matching
- Simple read commands return JSON via `json.NewEncoder(os.Stdout)`
- The `cmd_query.go` file shows the pattern for JSON output commands
- The `cmd_answer.go` file shows how to access `t.PlanQuestions`

## Approach
Create a new command file `cmd_questions.go` following the pattern from `cmd_query.go`. The command will load the ticket, extract the `PlanQuestions` field, and marshal it to JSON. If no questions exist, output an empty array `[]`.

## Tasks
1. [cmd_questions.go] — Create new file with `cmdQuestions(args)` function.
   - Accept ticket ID as single argument
   - Use `resolveProjectTicketsDir()` for project resolution
   - Use `ResolveID()` for partial ID matching
   - Load ticket with `LoadTicket()`
   - Marshal `t.PlanQuestions` to JSON using `json.MarshalIndent()` for readability
   - Output to stdout
   - Return 0 on success, 1 on error
   Verify: Command compiles without errors.

2. [main.go:run] — Add `"questions"` case to switch statement that calls `cmdQuestions(rest)`.
   - Insert after line 58 (after `"answer"` case) to maintain alphabetical ordering
   Verify: `go build` succeeds, `./ko questions` shows usage error.

3. [main.go:cmdHelp] — Add `questions <id>` to help text.
   - Insert after the `answer` command line for consistency with ordering
   - Format: `  questions <id>     Show plan questions as JSON`
   Verify: `ko help` shows the new command in output.

4. [cmd_questions_test.go] — Create unit tests following pattern from `cmd_answer_test.go`.
   - Test with ticket that has plan-questions (should output JSON array)
   - Test with ticket that has no plan-questions (should output `[]`)
   - Test with nonexistent ticket ID (should error)
   - Test with partial ticket ID resolution
   Verify: `go test -run TestCmdQuestions` passes.

## Open Questions
None. The schema is well-defined, the command pattern is established, and the output format is specified in the ticket. The implementation is straightforward JSON marshaling with no architectural decisions required.
