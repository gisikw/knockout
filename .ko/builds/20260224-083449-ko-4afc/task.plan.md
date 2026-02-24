Perfect! I've written the implementation plan to the artifact directory. The plan covers:

## Summary

**Goal**: Extend the `ko blocked` command to show unblock reasons for blocked tickets.

**Key findings from the codebase**:
- `ko blocked` already exists but only lists tickets with unresolved dependencies
- Block reasons are stored as notes in the ticket body (e.g., "ko: FAIL at node 'X' — reason")
- Two types of blocked tickets exist: those with `status: blocked` and those with unresolved deps

**Approach**: 
- Add optional `<id>` argument to show a specific ticket's block reason
- Extract reasons by parsing the notes section for "ko: FAIL" or "ko: BLOCKED" markers
- When listing all blocked tickets, show both status=blocked tickets (with reasons) and dependency-blocked tickets

**5 tasks planned**:
1. Extend `cmdBlocked` to handle optional ID argument
2. Add `ExtractBlockReason()` helper function to parse notes
3. Update output format to include reasons
4. Update help text
5. Add spec scenarios

**Open question**: Should we add a structured `block_reason` field to frontmatter? The plan proposes using notes-based extraction as an MVP, deferring structured frontmatter to future work if needed.

The plan is ready for review and approval before implementation begins.
