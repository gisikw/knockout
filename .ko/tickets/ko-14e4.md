---
id: ko-14e4
status: open
deps: [ko-6733]
created: 2026-02-28T07:40:29Z
type: task
priority: 2
---
# ko agent start should run ko agent triage on all triageable tickets before starting work on ready tickets

## Notes

**2026-02-28 14:52:19 UTC:** Question: When should the loop honor a stop signal during triage?
Answer: Check between tickets (Recommended)
Honor stop signals only between triage runs, allowing current triage to complete

**2026-02-28 14:52:19 UTC:** Question: How should the loop handle triage failures?
Answer: Log and continue (Recommended)
Log triage failures but continue processing remaining tickets, consistent with how FAIL outcomes work

**2026-02-28 14:52:19 UTC:** Question: Should triaged tickets count toward the --max-tickets limit?
Answer: Do not count (Recommended)
Triage is housekeeping; only ready-to-build tickets consume the limit
