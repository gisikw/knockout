Plan written to `/home/dev/Projects/knockout/.ko/tickets/ko-860e.artifacts/plan.md`. The plan consolidates three existing commands (`ko block`, `ko questions`, `ko answer`) into a single `ko triage` command with a flag-based interface.

**Key implementation points:**
- Creates new `cmd_triage.go` with consolidated logic from the three existing commands
- Reuses existing patterns (flag parsing, mutation events, ticket validation)
- Maintains backward-compatible data model (PlanQuestions, status transitions)
- Tests follow existing test patterns from the removed command test files
- Removes old commands entirely (breaking change)

**Open questions flagged:**
1. Output format for bare `ko triage <id>` (human-readable vs JSON-only)
2. Note format for `--block` reason (should match `ExtractBlockReason` parser)
3. Whether to maintain aliases for backward compatibility (plan assumes hard removal)

The plan is ready for review.
