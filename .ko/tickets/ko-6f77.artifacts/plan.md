## Goal
Add --project flag to commands that don't receive ticket IDs, and remove hashtag-based project routing from `ko add`.

## Context
Currently, `resolveProjectTicketsDir()` (cmd_list.go:57) parses a `#project` hashtag from args to route commands to different project directories. It's called by:
- `cmdLs` (cmd_list.go:93)
- `cmdReady` (cmd_list.go:182)
- `cmdAgentLoop` (cmd_loop.go:84)
- `cmdAgentStart` (cmd_agent.go:104)
- `cmdAgentStop` (cmd_agent.go:173)
- `cmdAgentStatus` (cmd_agent.go:259)
- `cmdAgentInit` (cmd_build_init.go:12)

Separately, `cmdCreate` (cmd_create.go) uses `RouteTicket()` (registry.go:303) which parses hashtags from the title via `ParseTags()` (registry.go:284) to automatically route tickets to other projects.

The ticket requests:
1. Change `resolveProjectTicketsDir()` to accept a `--project` flag instead of parsing hashtags
2. Strip the hashtag parsing from `ko add` — use `--project` exclusively for routing

This is a UX change: explicit `--project` flag replaces implicit `#tag` inference in both contexts.

## Approach
Modify `resolveProjectTicketsDir()` to parse a `--project` flag (instead of hashtags) and return it alongside the tickets directory. Update all callers to pass args through flag parsing that includes `--project`. For `ko add`, stop calling `RouteTicket()` and instead use the explicit `--project` flag value (if provided) to determine the target project. If no `--project` is given, default to local project.

## Tasks
1. [cmd_list.go:resolveProjectTicketsDir] — Change to parse `--project` flag instead of hashtags. Accept args, parse with a flag.FlagSet, extract the project tag, look it up in the registry, return tickets dir + remaining args (without the flag). If no `--project`, fall back to local tickets dir.
   Verify: `go test ./... -run TestCmdLs` passes (if test exists).

2. [cmd_list.go:cmdLs] — Add `--project` to the reorderArgs map before calling resolveProjectTicketsDir. Ensure the flag is parsed early so resolveProjectTicketsDir can consume it.
   Verify: `ko ls --project=<tag>` works, `ko ls` without flag works.

3. [cmd_list.go:cmdReady] — Add `--project` to the reorderArgs map.
   Verify: `ko ready --project=<tag>` works.

4. [cmd_loop.go:cmdAgentLoop] — Add `--project` to the reorderArgs map.
   Verify: `ko agent loop --project=<tag>` works.

5. [cmd_agent.go:cmdAgentStart] — Pass args through resolveProjectTicketsDir as-is (it now handles flag parsing internally).
   Verify: `ko agent start --project=<tag>` works.

6. [cmd_agent.go:cmdAgentStop] — Pass args through resolveProjectTicketsDir.
   Verify: `ko agent stop --project=<tag>` works.

7. [cmd_agent.go:cmdAgentStatus] — Pass args through resolveProjectTicketsDir.
   Verify: `ko agent status --project=<tag>` works.

8. [cmd_build_init.go:cmdAgentInit] — Pass args through resolveProjectTicketsDir (currently ignores error, keep that pattern).
   Verify: `ko agent init --project=<tag>` works.

9. [cmd_create.go:cmdCreate] — Remove call to `RouteTicket()`. Parse a `--project` flag instead. If provided, look up the project in the registry and route to that tickets dir. If not provided, use local tickets dir. Remove hashtag parsing from title. No longer create audit tickets for routed tickets — just create the ticket in the target location and print its ID.
   Verify: `ko add --project=<tag> "title"` creates ticket in target project, `ko add "title #tag"` creates ticket locally with title including "#tag" literally.

10. [registry.go:RouteTicket] — Mark as deprecated in a comment (needed for backward compatibility but no longer used by ko add). Do not delete — other code may depend on it.
    Verify: No action needed, just documentation.

11. Write tests for --project flag behavior in cmd_list_test.go (or create if missing). Test that `--project` routes correctly, that missing projects error, and that local fallback works.
    Verify: New tests pass.

## Open Questions
None. The implementation is straightforward: replace hashtag parsing with explicit flag parsing in two independent subsystems (command routing and ticket creation routing).
