---
id: ko-e33d
status: open
deps: [ko-6733]
created: 2026-02-28T07:40:29Z
type: task
priority: 2
---
# If pipeline.yml has auto_triage: true, ko agent triage should be triggered automatically when a ticket is created or modified to have triage instructions

## Notes

**2026-02-28 14:52:49 UTC:** Question: Should auto-triage failure prevent the ticket from being created or updated?
Answer: Non-fatal (Recommended)
Ticket saves with triage field set; user sees warning and can retry manually. Safer because the ticket isn't lost.
