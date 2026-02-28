---
id: ko-6733
status: open
deps: [ko-13af, ko-68c3]
created: 2026-02-28T07:40:28Z
type: task
priority: 2
---
# ko agent triage <id> — load default model and prompt it to handle the ticket per its triage instructions. Provide ko usage tokens so the model can operate on tickets. On success (no error), clear the triage field from the ticket (if it still exists).

## Notes

**2026-02-28 13:45:29 UTC:** Question: Should `ko agent triage` require a pipeline configuration to exist, or fall back to a default adapter?
Answer: Require pipeline config
Fail with a clear error if no pipeline exists—ensures intentional setup

**2026-02-28 13:45:29 UTC:** Question: Should `ko agent triage` enforce `allowAll=true` for tool calls, or respect the pipeline configuration?
Answer: Force allowAll=true (Recommended)
Always enable all tool calls for triage operations, regardless of pipeline config—more reliable for triage work
