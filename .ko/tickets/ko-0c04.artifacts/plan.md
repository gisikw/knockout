## Goal
Add `plan-questions` field to the JSON output of `ko show [id] --json` to match the behavior of `ko ls --json`.

## Context
The codebase has two different JSON struct types for ticket serialization:
- `ticketJSON` in cmd_list.go (line 13-28) includes `PlanQuestions []PlanQuestion` field
- `showJSON` in cmd_show.go (line 12-29) lacks the `PlanQuestions` field

The `PlanQuestion` type is already defined in ticket.go (lines 45-58) and is properly parsed/serialized as part of the `Ticket` struct (line 35). The `ko ls --json` command uses `ticketToJSON()` helper (lines 31-52) which includes plan questions, while `ko show --json` constructs its JSON manually and omits them.

Recent commit (327bace) shows JSON output was recently modified ("Downcase json output"), suggesting JSON fields are actively maintained.

## Approach
Add the `PlanQuestions` field to the `showJSON` struct and populate it from the ticket when building the JSON response. This is a straightforward single-field addition that maintains consistency between the two JSON output formats.

## Tasks
1. [cmd_show.go:showJSON] — Add `PlanQuestions []PlanQuestion` field after line 28, before the `Body` field, with proper json tag `json:"plan-questions,omitempty"` to match the YAML field name convention.
   Verify: Field appears in struct definition.

2. [cmd_show.go:cmdShow] — Populate the `PlanQuestions` field in the JSON construction at line 91 by adding `PlanQuestions: t.PlanQuestions,` to the struct literal.
   Verify: `go build .` compiles successfully.

3. Test manually — Create or use an existing ticket with plan questions, run `ko show [id] --json` and verify the output includes the `plan-questions` field in the JSON.
   Verify: JSON output includes `plan-questions` array when questions exist; field is omitted when empty (due to `omitempty`).

## Open Questions
None — the implementation is straightforward. The field name convention follows existing patterns (`plan-questions` with hyphen matches YAML frontmatter), the type is already defined and used elsewhere, and `omitempty` ensures backward compatibility.
