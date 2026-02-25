---
id: ko-30ec
status: open
deps: []
created: 2026-02-25T13:00:14Z
type: task
priority: 2
---
# ko serve: scope POST /ko endpoint to a project

The POST /ko endpoint currently runs commands against whatever project ko
resolves from cwd. The subscribe endpoint already accepts a project param.

Add an optional "project" field to the JSON body:

    {"argv": ["ls", "--json"], "project": "#knockout"}

If present, resolve via registry (for #tags) or use as absolute path, then
set the working directory for the exec. If absent, use cwd.

This matches the subscribe endpoint's project resolution pattern and prevents
the server from leaking access to unintended projects.
