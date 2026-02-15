---
id: ko-c0c9
status: closed
deps: []
links: []
created: 2026-02-15T05:56:43Z
type: task
priority: 3
---
# Track ticket prefixes in project registry

The registry should store known prefixes for each project (e.g. fort-nix: {path: /home/dev/Projects/fort-nix, prefix: fn}). On register, detect prefix if tickets exist. On any registry load, lazily backfill prefixes for projects that have concretized tickets but no prefix yet. This enables efficient cross-project dep resolution: parse the prefix from the dep ID and look up directly instead of scanning all projects.
