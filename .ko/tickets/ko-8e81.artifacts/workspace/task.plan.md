Perfect! The plan has been written to `/home/dev/Projects/knockout/.ko/tickets/ko-8e81.artifacts/plan.md`. 

The plan covers:

**Goal**: Clear one-sentence summary of what this accomplishes.

**Context**: Key findings about the codebase structure, patterns, and constraints:
- Command registration pattern in main.go
- File naming conventions (cmd_*.go)
- Testing patterns (testscript)
- No existing HTTP server
- INVARIANTS.md requirements

**Approach**: High-level strategy for implementation - new cmd_serve.go file with HTTP server, JSON endpoint, whitelist validation, and signal handling.

**Tasks**: Six ordered implementation steps:
1. Create cmd_serve.go with HTTP server logic
2. Register command in main.go switch
3. Add to help text
4. Create unit tests
5. Create integration tests
6. Register test function

**Open Questions**: Four specific questions about implementation choices:
1. Whether to include create/add in whitelist (decided to exclude per explicit ticket list)
2. Logging strategy for requests and errors
3. Content-Type validation approach
4. Working directory for exec.Command

Each task includes verification steps and the plan stays focused on what needs to be done without including actual code.
