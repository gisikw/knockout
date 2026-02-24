---
id: ko-72cd
status: open
deps: [ko-dd51]
created: 2026-02-24T12:49:29Z
type: task
priority: 2
---
# Pipeline support for resolved status: configure success outcome per workflow

Once `resolved` status exists (ko-dd51), the pipeline needs a way to
land tickets there instead of `closed` on success.

Options:
- New disposition type (`resolved` alongside `continue`/`fail`/etc.)
- Workflow-level config: `on_success: resolved` vs `on_success: closed`
- Convention: research and bug-wontfix workflows default to `resolved`,
  task workflows default to `closed`

The right answer probably involves both — a yaml-level default per
workflow, plus a disposition that lets individual decision nodes override.
Research workflows should land in `resolved` by default so findings
get human review before closing.
