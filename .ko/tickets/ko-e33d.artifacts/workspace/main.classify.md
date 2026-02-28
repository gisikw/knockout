```json
{"disposition": "route", "workflow": "task"}
```

This ticket describes a feature to implement: automatically trigger `ko agent triage` when a ticket is created or modified with triage instructions, if `pipeline.yml` has `auto_triage: true` enabled. The expected output is committed code changes to add this conditional automation. The notes provide implementation guidance (non-fatal failure handling). This is a clear task.
