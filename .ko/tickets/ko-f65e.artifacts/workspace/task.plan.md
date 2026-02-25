Perfect! I've created a comprehensive plan for consolidating the `.ko/` configuration files. The plan includes:

## Summary

The plan consolidates `.ko/pipeline.yml` and `.ko/prefix` into a unified `.ko/config.yaml` file with two top-level sections:
- `project:` - for project-level settings like prefix (and future settings like default priority, harness overrides)
- `pipeline:` - containing the current pipeline configuration

**Key design decisions:**
1. **Backwards compatibility** - The system will check for `config.yaml` first, then fall back to reading `pipeline.yml` + `prefix` separately. This allows gradual migration without breaking existing setups.

2. **Minimal scope** - Only including `prefix` in the initial `project:` section. Other candidates mentioned in the ticket (default priority, agent harness overrides) are deferred to follow-up tickets once their requirements are clearer.

3. **No migration command** - Users migrate naturally when they run `ko agent init` or manually update configs. Transparent fallback is sufficient.

The plan breaks down into 9 concrete tasks covering:
- Config loading and parsing logic
- Prefix reading/writing functions  
- Command updates (init, agent init, build)
- Test updates
- Documentation updates

Each task has clear verification steps to ensure correctness.
