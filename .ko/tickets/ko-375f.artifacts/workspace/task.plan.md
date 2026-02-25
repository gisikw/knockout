Perfect! I've updated the plan to incorporate the latest ticket note that clarifies the prompt passing mechanism. The plan is now complete and ready for review.

## Summary

I've updated the existing plan at `.ko/tickets/ko-375f.artifacts/plan.md` to incorporate the newest context from the ticket notes. The key update is that the open question about prompt passing has been resolved:

**Resolution**: Shell harnesses receive the prompt via `KO_PROMPT` environment variable only, and the scripts decide how to pass it to their agent (stdin, argument, etc.).

The plan is comprehensive and covers:

1. **Goal**: Replace YAML harness templates with shell scripts using KO_-namespaced env vars
2. **Context**: Detailed analysis of current YAML architecture, test patterns, and project constraints
3. **Approach**: Shell scripts receive KO_PROMPT, KO_MODEL, KO_SYSTEM_PROMPT, KO_ALLOW_ALL, KO_ALLOWED_TOOLS via environment and handle their own binary fallback and argument construction
4. **Tasks**: 17 ordered tasks covering:
   - Creating specs and shell scripts for claude/cursor harnesses
   - Implementing ShellAdapter in Go
   - Migrating tests to shell-based approach
   - Removing YAML support entirely
   - Updating documentation

All open questions have been resolved with answers from the ticket notes.
