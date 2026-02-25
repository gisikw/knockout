---
id: ko-f5e6
status: closed
deps: []
created: 2026-02-25T19:28:32Z
type: task
priority: 2
---
# Remove ko blocked command, fold block-reason into ko show

## Notes

**2026-02-25 19:30:04 UTC:** ko blocked without args duplicates ko ls --status=blocked. The single-ticket block-reason view (ko blocked <id>) should move to ko show output. Remove the command and update serve.go whitelist.

**2026-02-25 20:11:27 UTC:** Superseded by ko-860e (ko triage consolidation) which replaces block entirely
