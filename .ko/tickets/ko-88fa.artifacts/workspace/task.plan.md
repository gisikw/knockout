Perfect! I've created a comprehensive plan based on my investigation of the codebase. Here's what I found:

## Summary

The main issue is in **cmd_serve.go** - the HTTP server's whitelist still includes the old `query` command, which should be removed since it was replaced by `ls --json` in the CLI simplification.

The good news is:
- The prompt files in `.ko/prompts/` don't hardcode any CLI command invocations
- The build pipeline in `build.go` doesn't hardcode ko commands internally  
- Most old command names (`create`, `add-note`, `reopen`, `init`, `register`, `projects`) are already correctly excluded from the whitelist

The plan outlines 5 simple tasks to:
1. Remove `query` from the serve whitelist
2. Update the corresponding test whitelists
3. Update test expectations
4. Clarify a comment about excluded commands
5. Verify all tests pass

This is a focused, surgical change with clear verification steps at each stage.
