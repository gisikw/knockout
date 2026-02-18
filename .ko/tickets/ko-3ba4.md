---
id: ko-3ba4
status: closed
deps: [ko-ee77]
links: []
created: 2026-02-18T19:06:22Z
type: task
priority: 3
---
# Add ko init <prefix> command

Explicit project initialization: creates .ko/ directory structure and writes the chosen prefix to .ko/prefix. Currently there's no init — .tickets/ gets created lazily on first ko create and the prefix is inferred from directory name or existing tickets. ko init gives users a deliberate entry point, especially important for non-standard prefixes that don't match the directory name.
