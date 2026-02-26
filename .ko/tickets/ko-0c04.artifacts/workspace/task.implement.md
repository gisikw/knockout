**Task 3 verification:** ✅ Manual testing confirms:
- JSON output includes `plan-questions` array when questions exist (verified with ko-test1)
- Field is omitted when empty due to `omitempty` tag (verified with ko-0c04)

---

## Summary

I successfully implemented the ticket by following the plan exactly:

**Task 1 [cmd_show.go:12-30]** — Added `PlanQuestions []PlanQuestion` field to the `showJSON` struct with the json tag `json:"plan-questions,omitempty"` to match the convention used in `ko ls`.

**Task 2 [cmd_show.go:76-94]** — Populated the `PlanQuestions` field in the JSON construction by adding `PlanQuestions: t.PlanQuestions,` to the struct literal at line 92. The code compiled successfully.

**Task 3** — Manual testing confirmed:
- Created a test ticket with plan-questions in the frontmatter
- Verified `ko show ko-test1 --json` includes the `plan-questions` field with proper nesting and lowercase field names
- Verified `ko show ko-0c04 --json` omits the field when no questions exist (due to `omitempty`)
- The output format matches `ko ls --json` behavior

**No deviations from the plan.** The implementation was straightforward as described, and all verification steps passed.
