---
id: ko-88fa
status: open
deps: [ko-1bfc, ko-b99d, ko-dbe4, ko-70b8, ko-5734, ko-7537, ko-860e]
created: 2026-02-25T20:14:33Z
type: task
priority: 2
---
# Update ko agent build prompts/pipeline to use new CLI syntax

After CLI simplification, the agent build pipeline may use old command names in prompts or internal dispatch. Audit and update:
- .ko/prompts/ templates
- cmd_build.go internal command calls
- cmd_serve.go POST /ko whitelist
- Any hardcoded command names in pipeline execution

Old -> New:
- create -> add
- add-note -> note
- reopen -> open
- blocked/closed/resolved -> ls --status=X
- query -> ls --json
- block/questions/answer -> triage
- init/register/default -> project set
- projects -> project ls
