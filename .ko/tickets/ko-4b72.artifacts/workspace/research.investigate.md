## Investigation Complete

I've written comprehensive findings to `.ko/tickets/ko-4b72.artifacts/findings.md`.

### Key Findings:

**Yes, we can use ticket titles in commit messages.** The infrastructure is already there - we just need to expose it.

**Currently exposed environment variables:**
- `TICKET_ID` - ticket identifier (e.g., "ko-4b72")
- `CHANGED_FILES` - space-separated list of modified files  
- `KO_TICKET_WORKSPACE` - build workspace directory
- `KO_ARTIFACT_DIR` - artifacts directory
- `KO_BUILD_HISTORY` - build history log path

**What's needed:**
The `Title` field exists in the `Ticket` struct and is already populated (from the first `# ` heading in the markdown), but it's not exposed as an environment variable. Adding `TICKET_TITLE` requires just 2 lines in `build.go:runHooks()` plus documentation updates.

**Example result:**
Instead of `ko: ko-0c04`, commits could be:
```
Ensure ko show [id] --json includes questions json

ko: ko-0c04
```

The findings document includes implementation details, edge cases to consider (empty titles, long titles, special characters), and specific file/line references for making the changes.
