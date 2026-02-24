## Summary

I've completed the research investigation and documented my findings in `findings.md`.

**Answer: Yes, `agent.lock` and `agent.pid` (and also `agent.log`) should be gitignored, and `ko agent init` should write a `.ko/.gitignore` to handle this.**

### Key Findings:

1. **Current Problem**: These files are currently tracked by git, but they represent ephemeral runtime state that shouldn't be committed:
   - `agent.lock` - flock mechanism for preventing duplicate agents
   - `agent.pid` - process ID of the background agent
   - `agent.log` - runtime log output (currently showing as modified)

2. **Existing Precedent**: Commit `3b239b8` already established this pattern by adding `.ko/builds/.gitignore` to ignore build artifacts during `ko agent init`

3. **Classification**: Ticket ko-f65e explicitly classified these as "runtime state files" distinct from configuration

4. **Standard Practice**: `.pid`, `.lock`, and `.log` files are universally gitignored as they represent machine-specific process state

### Recommended Implementation:

Add to `cmd_build_init.go:cmdAgentInit()` (after line 44):
```go
// Ensure .ko/.gitignore covers runtime files
koGitignore := filepath.Join(koDir, ".gitignore")
runtimeIgnores := []string{"agent.lock", "agent.pid", "agent.log", ""}
os.WriteFile(koGitignore, []byte(strings.Join(runtimeIgnores, "\n")), 0644)
```

This follows the same pattern as the existing builds/.gitignore scaffolding.
