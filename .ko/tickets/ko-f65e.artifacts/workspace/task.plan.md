I've created a comprehensive plan for consolidating `.ko/` top-level files into a unified `config.yaml`. The plan addresses:

**Goal**: Merge `pipeline.yml` and the `prefix` file into a single `.ko/config.yaml` with clear sections.

**Key Context**: 
- Found 6 top-level files in `.ko/`, but only 2 need consolidation (pipeline.yml and prefix)
- Runtime files (lock, pid, log, heartbeat) remain separate as they're gitignored
- Custom minimal YAML parser approach should be maintained (no external deps)
- Multiple call sites across cmd_agent.go, cmd_build.go, cmd_loop.go need updates

**Approach**: 
- Create new `ProjectConfig` struct wrapping `Pipeline` + prefix
- Backward compatibility: try `config.yaml` first, fall back to `pipeline.yml`
- Update both read paths and write paths (especially prefix operations)
- Migrate examples and knockout's own config

**Tasks**: 9 ordered steps covering:
1. New data structure and parser
2. Tests for new parser
3. Config discovery updates
4. Pipeline loading updates  
5. Prefix read/write updates
6. Init command updates
7. Example migrations
8. Self-hosting migration
9. Documentation

**Open Questions**:
1. Whether to add `default_priority` now (mentioned in ticket but not implemented)
2. Deprecation timeline for pipeline.yml vs indefinite dual support
3. Whether `ko init` should also write config.yaml or stay focused on prefix only
