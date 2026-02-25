---
id: ko-7537
status: open
deps: []
created: 2026-02-25T20:06:42Z
type: task
priority: 2
---
# Consolidate init/register/default/projects into ko project {set,ls}

Replace four commands with two:

ko project set #tag [--prefix=pre] [--default]
  - Upsert: inits .ko dir if needed, registers in global registry, optionally sets default
  - Replaces: ko init, ko register, ko default (setter)

ko project ls
  - Lists registered projects with default marker
  - Replaces: ko projects, ko default (getter — default visible in ls output)

Remove: ko init, ko register, ko default, ko projects
