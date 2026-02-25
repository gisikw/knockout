Perfect! I've completed the plan for implementing the `ko questions` command. The plan is now written to `/home/dev/Projects/knockout/.ko/tickets/ko-3d1f.artifacts/plan.md`.

## Summary

The plan outlines a straightforward implementation:

**Goal**: Add `ko questions <id>` to read plan-questions from ticket frontmatter and output as JSON.

**Key findings**:
- The plan-questions schema already exists in the `Ticket` struct (from ko-9c4a)
- The codebase has clear patterns for JSON output commands (`cmd_query.go`)
- The implementation follows established conventions for command structure

**Implementation approach**:
1. Create `cmd_questions.go` with the new command following the `cmd_query.go` pattern
2. Wire it into `main.go`'s command router
3. Add help text
4. Add unit tests following the `cmd_answer_test.go` pattern

**No open questions** — the schema is defined, the patterns are clear, and the requirement is unambiguous.
