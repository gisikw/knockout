Perfect! I've created a comprehensive plan for consolidating the `.ko/` configuration files. The plan:

**Goal**: Merge `pipeline.yml` and `prefix` into a single `.ko/config.yaml`

**Key decisions**:
- Runtime state files (`agent.lock`, `agent.pid`, `agent.log`) stay separate — they're not config
- Extend the existing `Pipeline` struct and parser to handle the new `prefix` field
- Use the same minimal YAML parsing approach (no external dependencies)
- Provide a migration path: check for `config.yaml` first, fall back to old files

**Approach**: Add a `Prefix` field to the `Pipeline` struct, extend the parser to recognize it, and update all code paths that read/write these configs. The work is split into 14 tasks covering parsing, I/O, tests, examples, and documentation.

**Open Questions**: I've identified three architectural decisions:
1. Whether to support backwards compatibility with `pipeline.yml` (recommended: yes, with deprecation warning)
2. How to handle existing `.ko/prefix` files (recommended: read fallback, no dual-write)
3. Whether prefix should remain writable post-init (recommended: yes, keep existing behavior)

The plan is proportional to the task, includes concrete verification steps for each change, and maintains the pure-function/I/O-separation pattern from INVARIANTS.md.
