---
id: ko-6612
status: open
deps: []
created: 2026-02-25T23:29:16Z
type: task
priority: 2
---
# Implement ko update command

Universal ticket mutation command. Usage:

ko update <id> [--title title] [-d description] [-t type] [-p priority] [-a assignee]
               [--parent id] [--external-ref ref]
               [--design notes] [--acceptance criteria]
               [--tags tag1,tag2]
               [--questions json] [--answers json]
               [--status status]

Tags replace, not append. Should handle auto-unblocking: if ticket is blocked with questions and --answers resolves all open questions, transition to open.
