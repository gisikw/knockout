---
id: ko-5b3c
status: blocked
deps: [ko-bfe7]
created: 2026-03-01T06:18:30Z
type: task
priority: 2
plan-questions:
  - id: q1
    question: "Should hidden projects be revealed with a `--all` flag, or remain permanently unlisted?"
    context: "The plan excludes hidden projects from `ko project ls` output. If users can't discover what they've hidden, they may lose track of hidden projects. A `--all` flag provides an escape hatch for visibility."
    options:
      - label: "Always hidden (Recommended)"
        value: always_hidden
        description: "Hidden projects never appear in listings; users must know the tag to reference them"
      - label: "Reveal with --all flag"
        value: all_flag
        description: "Add `ko project ls --all` to show hidden projects alongside normal ones"
---
# Add support for a "hidden" flag in projects (which can be set similarly to the --default flag). If a project is hidden, it should be excluded from ko project ls
