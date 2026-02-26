## Summary

I've completed the investigation and documented my findings in `findings.md`. Here's what I discovered:

### Key Finding: The Plan Questions Workflow Works Correctly ✅

This ticket successfully tested the complete plan questions lifecycle:

1. **Plan generated with open questions** - The plan.md contained 3 genuine questions (Q1-Q3) about implementation approach
2. **Actionable gate converted them** - The actionable.md prompt detected the questions and converted them to structured JSON with IDs, question text, context, and options
3. **Build pipeline blocked appropriately** - The `needs_input` disposition set status to blocked and stored questions
4. **Answers were recorded** - Three timestamped notes document the chosen answers
5. **Workflow adapted** - On second build, ticket was correctly reclassified as research

### How It Works (Technical Details)

The flow involves:
- **disposition.go:16** - `PlanQuestions` field in Disposition struct
- **build.go:335-338** - Pipeline copies questions to ticket and blocks
- **ticket.go:35,117-134** - Serialization to/from YAML frontmatter  
- **cmd_update.go:147-215** - `ko update --answers` to provide responses
- **.ko/prompts/actionable.md** - Instructions for detecting and structuring questions

### Validation

All expected behaviors were confirmed working:
- Question detection and conversion
- Structured data validation (`ValidatePlanQuestions`)
- Ticket blocking mechanism
- Answer recording and auto-unblocking
- Build pipeline integration

The ticket's meta-goal—"I'd really like this ticket to end up with questions in its plan, as I'm trying to test that behavior"—was achieved. The feature is production-ready.
